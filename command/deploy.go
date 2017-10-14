package command

import (
	"fmt"
	"strings"

	nomad "github.com/hashicorp/nomad/api"

	"github.com/jrasell/levant/levant"
	"github.com/jrasell/levant/logging"
)

// DeployCommand is the command implementation that allows users to deploy a
// Nomad job based on passed templates and variables.
type DeployCommand struct {
	args []string
	Meta
}

// Help provides the help information for the deploy command.
func (c *DeployCommand) Help() string {
	helpText := `
Usage: levant deploy [options] TEMPLATE

  Deploy a Nomad job based on input templates and variable files.

General Options:

  -address=<http_address>
    The Nomad HTTP API address including port which Levant will use to make
    calls.

  -canary-auto-promote=<seconds>
    The time in seconds, after which Levant will auto-promote a canary job
    if all canaries within the deployment are healthy.

  -log-level=<level>
    Specify the verbosity level of Levant's logs. Valid values include DEBUG,
    INFO, and WARN, in decreasing order of verbosity. The default is INFO.

  -var-file=<file>
    Used in conjunction with the -job-file will deploy a templated job to your
    Nomad cluster.
`
	return strings.TrimSpace(helpText)
}

// Synopsis is provides a brief summary of the deploy command.
func (c *DeployCommand) Synopsis() string {
	return "Render and deploy a Nomad job from a template"
}

// Run triggers a run of the Levant template and deploy functions.
func (c *DeployCommand) Run(args []string) int {

	var variables, addr, log string
	var err error
	var job *nomad.Job
	var canary int

	flags := c.Meta.FlagSet("deploy", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.StringVar(&addr, "address", "", "")
	flags.IntVar(&canary, "canary-auto-promote", 0, "")
	flags.StringVar(&log, "log-level", "INFO", "")
	flags.StringVar(&variables, "var-file", "", "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

	if len(args) != 1 {
		c.UI.Error(c.Help())
		return 1
	}

	logging.SetLevel(log)

	job, err = levant.RenderJob(args[0], variables, &c.Meta.flagVars)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	if err = c.checkCanaryAutoPromote(job, canary); err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	client, err := levant.NewNomadClient(addr)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	success := client.Deploy(job, canary)
	if !success {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: deployment of job %s failed", *job.Name))
		return 1
	}

	c.UI.Info(fmt.Sprintf("[INFO] levant/command: deployment of job %s successful", *job.Name))

	return 0
}

func (c *DeployCommand) checkCanaryAutoPromote(job *nomad.Job, canaryAutoPromote int) error {

	if canaryAutoPromote > 0 && *job.Update.Canary == 0 {
		return fmt.Errorf("canary-auto-update of %v passed but job is not canary enabled", canaryAutoPromote)
	}

	c.UI.Info(fmt.Sprintf("[INFO] levant/command: running canary-auto-update of %v", canaryAutoPromote))

	return nil
}
