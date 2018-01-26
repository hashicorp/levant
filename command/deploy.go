package command

import (
	"fmt"
	"strings"

	nomad "github.com/hashicorp/nomad/api"

	"github.com/jrasell/levant/helper"
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
Usage: levant deploy [options] [TEMPLATE]

  Deploy a Nomad job based on input templates and variable files.

Arguments:

  TEMPLATE  nomad job template
    If no argument is given we look for a single *.nomad file

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
    Nomad cluster. [default: levant.(yaml|yml|tf)]

  -force-count
    Use the taskgroup count from the Nomad jobfile instead of the count that
    is currently set in a running job.

  -monitor
    Monitor the job until it completes.  Only applies to batch style jobs (since
    services don't complete).
	
  -timeout
    How long to wait ( in seconds ) for jobs to be successful before giving up. 
    Only applies to batch type jobs for now. Default is 300 seconds.
`
	return strings.TrimSpace(helpText)
}

// Synopsis is provides a brief summary of the deploy command.
func (c *DeployCommand) Synopsis() string {
	return "Render and deploy a Nomad job from a template"
}

// Run triggers a run of the Levant template and deploy functions.
func (c *DeployCommand) Run(args []string) int {

	var variables, addr, log, templateFile string
	var err error
	var job *nomad.Job
	var canary int
	var forceCount bool
	var monitor bool
	var timeout uint64

	flags := c.Meta.FlagSet("deploy", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.StringVar(&addr, "address", "", "")
	flags.IntVar(&canary, "canary-auto-promote", 0, "")
	flags.StringVar(&log, "log-level", "INFO", "")
	flags.StringVar(&variables, "var-file", "", "")
	flags.BoolVar(&forceCount, "force-count", false, "")
	flags.BoolVar(&monitor, "monitor", false, "")
	flags.Uint64Var(&timeout, "timeout", 300, "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

	logging.SetLevel(log)

	if len(args) == 1 {
		templateFile = args[0]
	} else if len(args) == 0 {
		if templateFile = helper.GetDefaultTmplFile(); templateFile == "" {
			c.UI.Error(c.Help())
			c.UI.Error("\nERROR: Template arg missing and no default template found")
			return 1
		}
	} else {
		c.UI.Error(c.Help())
		return 1
	}

	job, err = levant.RenderJob(templateFile, variables, &c.Meta.flagVars)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	if canary > 0 {
		if err = c.checkCanaryAutoPromote(job, canary); err != nil {
			c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
			return 1
		}

		c.UI.Info(fmt.Sprintf("[INFO] levant/command: running canary-auto-update of %vs", canary))
	}

	client, err := levant.NewNomadClient(addr)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	success := client.Deploy(job, canary, forceCount, timeout, monitor)
	if !success {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: deployment of job %s failed", *job.Name))
		return 1
	}

	c.UI.Info(fmt.Sprintf("[INFO] levant/command: deployment of job %s successful", *job.Name))

	return 0
}

func (c *DeployCommand) checkCanaryAutoPromote(job *nomad.Job, canaryAutoPromote int) error {
	if canaryAutoPromote == 0 {
		return nil
	}

	if job.Update != nil && job.Update.Canary != nil && *job.Update.Canary > 0 {
		return nil
	}

	for _, group := range job.TaskGroups {
		if group.Update != nil && group.Update.Canary != nil && *group.Update.Canary > 0 {
			return nil
		}
	}

	return fmt.Errorf("canary-auto-update of %v passed but job is not canary enabled", canaryAutoPromote)
}
