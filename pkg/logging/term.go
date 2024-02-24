package logging

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/numtide/nits/pkg/agent/info"
	"github.com/numtide/nits/pkg/subject"

	"github.com/charmbracelet/log"
	"github.com/nats-io/nats.go"
)

type TerminalRecord struct {
	msg       *nats.Msg
	agentInfo *info.Response
}

func (t *TerminalRecord) Type() RecordType {
	return RecordTerm
}

func (t *TerminalRecord) Msg() *nats.Msg {
	return t.msg
}

func (t *TerminalRecord) Write(file *os.File) (n int, err error) {
	b := bytes.NewBuffer(nil)

	var meta *nats.MsgMetadata
	if meta, err = t.msg.Metadata(); err != nil {
		return
	}

	styles := log.DefaultStyles()

	b.WriteString(styles.Timestamp.Render(meta.Timestamp.Format(time.RFC3339)))
	b.WriteByte(' ')

	// by default the prefix is just the msg subject
	prefix := t.msg.Subject
	if t.agentInfo != nil {
		prefix = fmt.Sprintf("%s | %s", t.agentInfo.Name, strings.TrimPrefix(t.msg.Subject, subject.AgentLogs(t.agentInfo.NKey)+"."))
	}

	b.WriteString(styles.Prefix.Render(prefix))
	b.WriteString("\n")
	b.WriteString(styles.Message.Render(string(t.msg.Data)))
	b.WriteByte(' ')

	b.WriteString("\n")
	return file.Write(b.Bytes())
}

func UnmarshalTerminalRecord(ctx context.Context, msg *nats.Msg, record *TerminalRecord) (err error) {
	if msg.Header.Get(HeaderFormat) != HeaderTerm {
		return ErrUnexpectedFormat
	}
	record.msg = msg

	// look up agent info based on the subject
	byNKey := GetAgentsByNKey(ctx)
	nkey := subject.AgentNKeyForSubject(msg.Subject)

	agentInfo, ok := byNKey[nkey]
	if ok {
		record.agentInfo = agentInfo
	}

	return
}
