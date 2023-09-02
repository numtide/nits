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

func ListWithContext(ctx context.Context, conn *nats.Conn) (agents []*info.Response, err error) {
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
