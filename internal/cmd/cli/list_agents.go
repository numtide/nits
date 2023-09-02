package cli

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	"github.com/numtide/nits/pkg/agent/info"
	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/xeonx/timeago"
)

type listAgentsCmd struct {
	Nats nutil.CliOptions `embed:"nats-"`
}

func (l *listAgentsCmd) Run() error {
	Cmd.Log.ConfigureLog()

	return cmd.Run(func(ctx context.Context) (err error) {
		var conn *nats.Conn
		if conn, err = l.Nats.Connect(); err != nil {
			return
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var agents []*info.Response
		if agents, err = agent.List(ctx, conn); err != nil {
			return err
		}

		columns := []table.Column{
			{Title: "Name", Width: 32},
			{Title: "NKey", Width: 57},
			{Title: "Last Seen", Width: 24},
		}

		var rows []table.Row
		for _, v := range agents {
			row := table.Row{v.Name, v.NKey, timeago.English.Format(v.LastSeen)}
			rows = append(rows, row)
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(false),
			table.WithHeight(len(rows)),
		)

		t.SetStyles(tableStyle)

		print(t.View())

		return
	})
}
