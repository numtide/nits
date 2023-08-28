package subject

import "fmt"

func AgentDeploymentWithNKey(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.DEPLOYMENT", Prefix, nkey)
}

func AgentDeploymentWithName(name string) string {
	return fmt.Sprintf("%s.AGENT.NAME.%s.DEPLOYMENT", Prefix, name)
}

func AgentLogs(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.LOGS", Prefix, nkey)
}

func AgentInbox(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.INBOX", Prefix, nkey)
}
