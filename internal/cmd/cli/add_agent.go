package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	nsccmd "github.com/nats-io/nsc/v2/cmd"

	"github.com/numtide/nits/internal/cmd"

	nexec "github.com/numtide/nits/pkg/exec"
	nutil "github.com/numtide/nits/pkg/nats"

	"github.com/numtide/nits/pkg/subject"

	"github.com/juju/errors"
	"golang.org/x/crypto/ssh"
)

type addAgentCmd struct {
	Cluster        string `required:"" help:"Name of the account under which Agents will run"`
	PublicKey      string `required:"" xor:"key"`
	PublicKeyFile  string `required:"" type:"existingfile" xor:"key"`
	PrivateKeyFile string `required:"" type:"existingfile" xor:"key"`

	Name string `arg:"" help:"A name for the agent account"`
}

func (a *addAgentCmd) Run() (err error) {
	Cmd.Log.ConfigureLog()

	var nkey string

	// todo move this logic somewhere shared
	if a.PrivateKeyFile != "" {
		var signer ssh.Signer
		if signer, err = nutil.NewSigner(a.PrivateKeyFile); err != nil {
			return errors.Annotate(err, "failed to parse private key file")
		} else if nkey, err = nutil.NKeyForSigner(signer); err != nil {
			return err
		}
	} else {

		var pk ssh.PublicKey
		keyBytes := []byte(a.PublicKey)

		if !(a.PublicKey == "" || strings.Contains(a.PublicKey, "ssh-ed25519")) {
			keyBytes = []byte("ed25519 " + a.PublicKey)
		} else if a.PublicKeyFile != "" {
			if keyBytes, err = os.ReadFile(a.PublicKeyFile); err != nil {
				return errors.Annotate(err, "failed to read public key file")
			}
		}

		if pk, _, _, _, err = ssh.ParseAuthorizedKey(keyBytes); err != nil {
			return errors.Annotate(err, "failed to parse public key")
		}

		nkey, err = nutil.NKeyForPublicKey(pk)
		if err != nil {
			return errors.Annotate(err, "failed to determine nkey for public key")
		}
	}

	var op nsccmd.OperatorDescriber
	if op, err = cmd.DetectOperator(); err != nil {
		return
	}

	agentSubject := fmt.Sprintf("NITS.AGENT.%s.>", nkey)
	agentByName := subject.AgentWithName(a.Name)
	agentInfoService := subject.AgentService(nkey, "INFO")

	log.Info("adding a subject mapping", "from", agentByName, "to", agentInfoService)

	nsc := cmd.LogExec(
		nexec.Nsc("add", "mapping", "-a", a.Cluster,
			"--from", agentByName,
			"--to", agentInfoService,
		),
	)

	if _, err = nsc.Output(); err != nil {
		nexec.LogError("failed to add subject mapping", err)
		return
	}

	log.Info("adding an agent user", "operator", op.Name, "account", a.Cluster, "name", a.Name)

	nsc = cmd.LogExec(
		nexec.Nsc(
			"add", "user", "-a", a.Cluster,
			"-k", nkey,
			"-n", a.Name,
			"--allow-pub", "NITS.CACHE.>",
			"--allow-pubsub", agentSubject,
			"--allow-pub", subject.AgentRegistration(nkey),
			"--allow-pub", "$JS.ACK.agent-deployments.>",
			"--allow-pub", "$JS.API.STREAM.NAMES",
			"--allow-pub", "$JS.API.CONSUMER.*.agent-deployments.>",
			"--allow-sub", "$SRV.>",
			"--allow-pub", "_INBOX.>",
		),
	)

	if _, err = nsc.Output(); err != nil {
		nexec.LogError("failed to add agent user", err)
		return
	}

	return
}
