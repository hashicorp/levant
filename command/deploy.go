// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/levant/helper"
	"github.com/hashicorp/levant/levant"
	"github.com/hashicorp/levant/levant/structs"
	"github.com/hashicorp/levant/logging"
	"github.com/hashicorp/levant/template"
	nomad "github.com/hashicorp/nomad/api"
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

  -allow-stale
    Allow stale consistency mode for requests into nomad.

  -canary-auto-promote=<seconds>
    The time in seconds, after which Levant will auto-promote a canary job
    if all canaries within the deployment are healthy.

  -consul-address=<addr>
    The Consul host and port to use when making Consul KeyValue lookups for
    template rendering.

  -force
    Execute deployment even though there were no changes.

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

  -vault
    This flag makes levant load the vault token from the current ENV.
    It can not be used at the same time than -vault-token=<vault-token> flag

  -vault-token=<vault-token>
    The vault token used to deploy the application to nomad with vault support
    This flag can not be used at the same time than -vault flag

  -log-level=<level>
    Specify the verbosity level of Levant's logs. Valid values include DEBUG,
    INFO, and WARN, in decreasing order of verbosity. The default is INFO.

  -log-format=<format>
    Specify the format of Levant's logs. Valid values are HUMAN or JSON. The
    default is HUMAN.

  -var-file=<file>
    Path to a file containing user variables used when rendering the job
    template. You can repeat this flag multiple times to supply multiple
    var-files. Defaults to levant.(json|yaml|yml|tf).
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
	var level, format string

	config := &levant.DeployConfig{
		Client:   &structs.ClientConfig{},
		Deploy:   &structs.DeployConfig{},
		Plan:     &structs.PlanConfig{},
		Template: &structs.TemplateConfig{},
	}

	flags := c.Meta.FlagSet("deploy", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.StringVar(&config.Client.Addr, "address", "", "")
	flags.BoolVar(&config.Client.AllowStale, "allow-stale", false, "")
	flags.IntVar(&config.Deploy.Canary, "canary-auto-promote", 0, "")
	flags.StringVar(&config.Client.ConsulAddr, "consul-address", "", "")
	flags.BoolVar(&config.Deploy.Force, "force", false, "")
	flags.BoolVar(&config.Deploy.ForceBatch, "force-batch", false, "")
	flags.BoolVar(&config.Deploy.ForceCount, "force-count", false, "")
	flags.BoolVar(&config.Plan.IgnoreNoChanges, "ignore-no-changes", false, "")
	flags.StringVar(&level, "log-level", "INFO", "")
	flags.StringVar(&format, "log-format", "HUMAN", "")
	flags.StringVar(&config.Deploy.VaultToken, "vault-token", "", "")
	flags.BoolVar(&config.Deploy.EnvVault, "vault", false, "")

	flags.Var((*helper.FlagStringSlice)(&config.Template.VariableFiles), "var-file", "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

	if config.Deploy.EnvVault && config.Deploy.VaultToken != "" {
		c.UI.Error(c.Help())
		c.UI.Error("\nERROR: Can not used -vault and -vault-token flag at the same time")
		return 1
	}

	if err = logging.SetupLogger(level, format); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if len(args) == 1 {
		config.Template.TemplateFile = args[0]
	} else if len(args) == 0 {
		if config.Template.TemplateFile = helper.GetDefaultTmplFile(); config.Template.TemplateFile == "" {
			c.UI.Error(c.Help())
			c.UI.Error("\nERROR: Template arg missing and no default template found")
			return 1
		}
	} else {
		c.UI.Error(c.Help())
		return 1
	}

	config.Template.Job, err = template.RenderJob(config.Template.TemplateFile,
		config.Template.VariableFiles, config.Client.ConsulAddr, &c.Meta.flagVars)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	if config.Deploy.Canary > 0 {
		if err = c.checkCanaryAutoPromote(config.Template.Job, config.Deploy.Canary); err != nil {
			c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
			return 1
		}
	}

	if config.Deploy.ForceBatch {
		if err = c.checkForceBatch(config.Template.Job, config.Deploy.ForceBatch); err != nil {
			c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
			return 1
		}
	}

	if !config.Deploy.Force {
		p := levant.PlanConfig{
			Client:   config.Client,
			Plan:     config.Plan,
			Template: config.Template,
		}

		planSuccess, changes := levant.TriggerPlan(&p)
		if !planSuccess {
			return 1
		} else if !changes && p.Plan.IgnoreNoChanges {
			return 0
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
