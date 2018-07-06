package command

import (
	"fmt"
	"strings"

	nomad "github.com/hashicorp/nomad/api"

	"github.com/jrasell/levant/helper"
	"github.com/jrasell/levant/levant"
	"github.com/jrasell/levant/levant/structs"
	"github.com/jrasell/levant/logging"
	"github.com/jrasell/levant/template"
)

// DeployCommand is the command implementation that allows users to deploy a
// Nomad job based on passed templates and variables.
type DeployCommand struct {
	Meta
}

// Help provides the help information for the deploy command.
func (c *DeployCommand) Help() string {
	helpText := `
Usage: levant deploy [options] [TEMPLATE]

  Deploy a Nomad job based on input templates and variable files. The deploy
  command supports passing variables individually on the command line. Multiple
  commands can be passed in the format of -var 'key=value'. Variables passed
  via the command line take precedence over the same variable declared within
  a passed variable file.

Arguments:

  TEMPLATE nomad job template
    If no argument is given we look for a single *.nomad file

General Options:

  -address=<http_address>
    The Nomad HTTP API address including port which Levant will use to make
    calls.

  -canary-auto-promote=<seconds>
    The time in seconds, after which Levant will auto-promote a canary job
    if all canaries within the deployment are healthy.
		
  -consul-address=<addr>
    The Consul host and port to use when making Consul KeyValue lookups for
    template rendering.

  -force-batch
    Forces a new instance of the periodic job. A new instance will be created
    even if it violates the job's prohibit_overlap settings.

  -force-count
    Use the taskgroup count from the Nomad jobfile instead of the count that
    is currently set in a running job.

  -ignore-no-changes
    By default if no changes are detected when running a deployment Levant will
    exit with a status 1 to indicate a deployment didn't happen. This behaviour
    can be changed using this flag so that Levant will exit cleanly ensuring CD
    pipelines don't fail when no changes are detected.

  -log-level=<level>
    Specify the verbosity level of Levant's logs. Valid values include DEBUG,
    INFO, and WARN, in decreasing order of verbosity. The default is INFO.

  -log-format=<format>
    Specify the format of Levant's logs. Valid values are HUMAN or JSON. The
    default is HUMAN.

  -var-file=<file>
    Used in conjunction with the -job-file will deploy a templated job to your
    Nomad cluster. You can repeat this flag multiple times to supply multiple var-files.
    [default: levant.(json|yaml|yml|tf)]
`
	return strings.TrimSpace(helpText)
}

// Synopsis is provides a brief summary of the deploy command.
func (c *DeployCommand) Synopsis() string {
	return "Render and deploy a Nomad job from a template"
}

// Run triggers a run of the Levant template and deploy functions.
func (c *DeployCommand) Run(args []string) int {

	var err error
	var addr string
	config := &structs.Config{}

	flags := c.Meta.FlagSet("deploy", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.StringVar(&config.Addr, "address", "", "")
	flags.IntVar(&config.Canary, "canary-auto-promote", 0, "")
	flags.StringVar(&addr, "consul-address", "", "")
	flags.BoolVar(&config.ForceBatch, "force-batch", false, "")
	flags.BoolVar(&config.ForceCount, "force-count", false, "")
	flags.BoolVar(&config.IgnoreNoChanges, "ignore-no-changes", false, "")
	flags.StringVar(&config.LogLevel, "log-level", "INFO", "")
	flags.StringVar(&config.LogFormat, "log-format", "HUMAN", "")
	flags.Var((*helper.FlagStringSlice)(&config.VariableFiles), "var-file", "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

	if err = logging.SetupLogger(config.LogLevel, config.LogFormat); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if len(args) == 1 {
		config.TemplateFile = args[0]
	} else if len(args) == 0 {
		if config.TemplateFile = helper.GetDefaultTmplFile(); config.TemplateFile == "" {
			c.UI.Error(c.Help())
			c.UI.Error("\nERROR: Template arg missing and no default template found")
			return 1
		}
	} else {
		c.UI.Error(c.Help())
		return 1
	}

	config.Job, err = template.RenderJob(config.TemplateFile, config.VariableFiles, addr, &c.Meta.flagVars)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	if config.Canary > 0 {
		if err = c.checkCanaryAutoPromote(config.Job, config.Canary); err != nil {
			c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
			return 1
		}
	}

	if config.ForceBatch {
		if err = c.checkForceBatch(config.Job, config.ForceBatch); err != nil {
			c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
			return 1
		}
	}

	success := levant.TriggerDeployment(config, nil)
	if !success {
		return 1
	}

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

// checkForceBatch ensures that if the force-batch flag is passed, the job is
// periodic.
func (c *DeployCommand) checkForceBatch(job *nomad.Job, forceBatch bool) error {

	if forceBatch && job.IsPeriodic() {
		return nil
	}

	return fmt.Errorf("force-batch passed but job is not periodic")
}
