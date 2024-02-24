package logging

import (
	"context"

	"github.com/numtide/nits/pkg/agent/info"
)

const (
	agentsByName = "agentsByName"
	agentsByNKey = "agentsByNKey"
)

type AgentIndex = map[string]*info.Response

func SetAgentsByName(ctx context.Context, byName AgentIndex) context.Context {
	return context.WithValue(ctx, agentsByName, byName)
}

func GetAgentsByName(ctx context.Context) AgentIndex {
	return ctx.Value(agentsByName).(AgentIndex)
}

func SetAgentsByNKey(ctx context.Context, byName AgentIndex) context.Context {
	return context.WithValue(ctx, agentsByNKey, byName)
}

func GetAgentsByNKey(ctx context.Context) AgentIndex {
	return ctx.Value(agentsByNKey).(AgentIndex)
}
