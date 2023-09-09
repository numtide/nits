package log

import (
	"time"

	"github.com/numtide/nits/pkg/agent/info"
	"github.com/numtide/nits/pkg/subject"

	"github.com/nats-io/nats.go"
)

const (
	HeaderEOF            = "EOF"
	HeaderEOFValue       = "End-Of-Stream"
	HeaderFormat         = "Format"
	HeaderFormatLogFmt   = "LogFmt"
	HeaderFormatTerminal = "Terminal"
	HeaderAgentName      = "AgentName"
	DefaultReadTimeout   = 1 * time.Second
)

type MsgProcessor func(msg *nats.Msg) error

func ProcessMsg(msg *nats.Msg, processors ...MsgProcessor) (err error) {
	for _, proc := range processors {
		if err = proc(msg); err != nil {
			return
		}
	}
	return
}

func GetAgentName(msg *nats.Msg) string {
	return msg.Header.Get(HeaderAgentName)
}

func ResolveAgentName(bySubject map[string]*info.Response) MsgProcessor {
	return func(msg *nats.Msg) error {
		if msg.Header == nil {
			msg.Header = make(nats.Header)
		}

		subject.AgentSubjectRegex()

		if agent, ok := bySubject[msg.Subject[:67]]; ok {
			msg.Header.Set(HeaderAgentName, agent.Name)
		}
		return nil
	}
}
