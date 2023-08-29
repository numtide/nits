package subject

import "fmt"

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

func AgentLogs(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.LOGS", Prefix, nkey)
}

func AgentService(nkey string, name string) string {
	return fmt.Sprintf("%s.AGENT.%s.SVC.%s", Prefix, nkey, name)
}

func AgentInbox(nkey string) string {
	return fmt.Sprintf("%s.AGENT.%s.INBOX", Prefix, nkey)
}
