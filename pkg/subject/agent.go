package subject

import (
	"fmt"
	"regexp"
)

func AgentSubjectRegex() *regexp.Regexp {
	return regexp.MustCompile("^(" + AgentPrefix() + ".\\w{56}).*")
}

func AgentWithNKey(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s", Prefix, nkey)
}

func AgentDeploymentWithNKey(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.DEPLOYMENT", Prefix, nkey)
}

func AgentWithName(name string) string {
	return fmt.Sprintf("%s.AGENT.NAME.%s", Prefix, name)
}

func AgentDeploymentWithName(name string) string {
	return fmt.Sprintf("%s.AGENT.NAME.%s.DEPLOYMENT", Prefix, name)
}

func AgentPrefix() string {
	return fmt.Sprintf("%s.AGENT", Prefix)
}

func AgentLogs(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.LOGS", Prefix, nkey)
}

func AgentLogsAll() string {
	return fmt.Sprintf("%s.AGENT.*.LOGS.>", Prefix)
}

func AgentService(nkey string, name string) string {
	return fmt.Sprintf("%s.AGENT.%s.SRV.%s", Prefix, nkey, name)
}

func AgentRegistry() string {
	return fmt.Sprintf("%s.AGENT_REGISTRY", Prefix)
}

func AgentRegistration(nkey string) string {
	return fmt.Sprintf("%s.%s", AgentRegistry(), nkey)
}

func AgentInbox(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.INBOX", Prefix, nkey)
}
