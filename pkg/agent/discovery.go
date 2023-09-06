package agent

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/charmbracelet/log"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/agent/info"
	"github.com/numtide/nits/pkg/subject"
)

func ResolveNKey(ctx context.Context, conn *nats.Conn, name string) (nkey string, err error) {
	var agents []*info.Response
	if agents, err = List(ctx, conn); err != nil {
		return
	}

	for _, a := range agents {
		if a.Name == name {
			nkey = a.NKey
			break
		}
	}

	if nkey == "" {
		return nkey, errors.Errorf("no agent found with name: %s", name)
	}

	return
}

func List(ctx context.Context, conn *nats.Conn) (agents []*info.Response, err error) {
	var (
		js   nats.JetStreamContext
		sub  *nats.Subscription
		msg  *nats.Msg
		meta *nats.MsgMetadata
	)

	if js, err = conn.JetStream(); err != nil {
		return
	} else if sub, err = js.SubscribeSync(subject.AgentRegistry()+".>", nats.DeliverAll()); err != nil {
		return
	}

	defer func() {
		if agents != nil {
			sort.SliceStable(agents, func(i, j int) bool {
				// descending order by last seen
				return agents[i].LastSeen.Compare(agents[j].LastSeen) >= 0
			})
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if msg, err = sub.NextMsgWithContext(ctx); !(err == nil || errors.Is(err, nats.ErrTimeout)) {
				return
			} else if msg != nil {
				var resp info.Response
				if err = json.Unmarshal(msg.Data, &resp); err != nil {
					// log the error but continue processing the remaining responses
					log.Error("failed to unmarshal agent info", "error", err)
				}

				if meta, err = msg.Metadata(); err != nil {
					return
				}

				resp.LastSeen = meta.Timestamp
				agents = append(agents, &resp)

				if meta.NumPending == 0 {
					// we have read everything in the stream
					return
				}
			}

		}
	}
}

func IndexByFunc(agents []*info.Response, keyFn func(*info.Response) (string, error)) (indexed map[string]*info.Response, err error) {
	var key string
	indexed = make(map[string]*info.Response, len(agents))
	for _, agent := range agents {
		if key, err = keyFn(agent); err != nil {
			return
		}
		indexed[key] = agent
	}
	return
}

func IndexByName(agents []*info.Response) (indexed map[string]*info.Response, err error) {
	return IndexByFunc(agents, func(response *info.Response) (string, error) {
		return response.Name, nil
	})
}

func IndexByNKey(agents []*info.Response) (indexed map[string]*info.Response, err error) {
	return IndexByFunc(agents, func(response *info.Response) (string, error) {
		return response.NKey, nil
	})
}

func IndexBySubject(agents []*info.Response) (indexed map[string]*info.Response, err error) {
	return IndexByFunc(agents, func(response *info.Response) (string, error) {
		return response.Subject, nil
	})
}

func ListByFunc(ctx context.Context, conn *nats.Conn, keyFn func(*info.Response) string) (agents map[string]*info.Response, err error) {
	var list []*info.Response
	if list, err = List(ctx, conn); err != nil {
		return
	}

	agents = make(map[string]*info.Response)
	for _, agent := range list {
		if _, ok := agents[agent.Name]; ok {
			return nil, errors.Errorf("more than one agent shares this name: %s", agent.Name)
		}
		agents[keyFn(agent)] = agent
	}
	return
}

func ListBySubject(ctx context.Context, conn *nats.Conn) (agents map[string]*info.Response, err error) {
	return ListByFunc(ctx, conn, func(agent *info.Response) string {
		return agent.Subject
	})
}
