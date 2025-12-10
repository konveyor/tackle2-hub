package ssh

import (
	"os"

	"github.com/konveyor/tackle2-hub/shared/addon/command"
	"github.com/konveyor/tackle2-hub/shared/addon/sink"
	"github.com/konveyor/tackle2-hub/shared/ssh"
)

func init() {
	ssh.Home, _ = os.Getwd()
	ssh.NewCommand = command.New
	ssh.Log = ssh.Log.WithSink(sink.New(true))
}

// Agent agent.
type Agent = ssh.Agent
