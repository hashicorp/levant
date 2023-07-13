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
)

// PlanCommand is the command implementation that allows users to plan a
// Nomad job based on passed templates and variables.
type PlanCommand struct {
	Meta
}

// Help provides the help information for the plan command.
func (c *PlanCommand) Help() string {
	helpText := `
Usage: levant plan [options] [TEMPLATE]

  Perform a Nomad plan based on input templates and variable files. The plan
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
		
  -consul-address=<addr>
    The Consul host and port to use when making Consul KeyValue lookups for
    template rendering.

  -force-count
    Use the taskgroup count from the Nomad jobfile instead of the count that
    is currently set in a running job.

  -ignore-no-changes
    By default if no changes are detected when running a plan Levant will
    exit with a status 1 to indicate there are no changes. This behaviour
    can be changed using this flag so that Levant will exit cleanly ensuring CD
    pipelines don't fail when no changes are detected.

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

// Synopsis is provides a brief summary of the plan command.
func (c *PlanCommand) Synopsis() string {
	return "Render and perform a Nomad job plan from a template"
}

// Run triggers a run of the Levant template and plan functions.
func (c *PlanCommand) Run(args []string) int {

	var err error
	var level, format string
	config := &levant.PlanConfig{
		Client:   &structs.ClientConfig{},
		Plan:     &structs.PlanConfig{},
		Template: &structs.TemplateConfig{},
	}

	flags := c.Meta.FlagSet("plan", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.StringVar(&config.Client.Addr, "address", "", "")
	flags.BoolVar(&config.Client.AllowStale, "allow-stale", false, "")
	flags.StringVar(&config.Client.ConsulAddr, "consul-address", "", "")
	flags.BoolVar(&config.Plan.IgnoreNoChanges, "ignore-no-changes", false, "")
	flags.StringVar(&level, "log-level", "INFO", "")
	flags.StringVar(&format, "log-format", "HUMAN", "")
	flags.Var((*helper.FlagStringSlice)(&config.Template.VariableFiles), "var-file", "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

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

	success, changes := levant.TriggerPlan(config)
	if !success {
		return 1
	} else if !changes && config.Plan.IgnoreNoChanges {
		return 0
	}

	return 0
}
