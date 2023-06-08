// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/levant/helper"
	"github.com/hashicorp/levant/logging"
	"github.com/hashicorp/levant/template"
)

// RenderCommand is the command implementation that allows users to render a
// Nomad job template based on passed templates and variables.
type RenderCommand struct {
	Meta
}

// Help provides the help information for the template command.
func (c *RenderCommand) Help() string {
	helpText := `
Usage: levant render [options] [TEMPLATE]

  Render a Nomad job template, useful for debugging. Like deploy, the render
  command also supports passing variables individually on the command line. 
  Multiple vars can be passed in the format of -var 'key=value'. Variables 
  passed via the command line take precedence over the same variable declared
  within a passed variable file.

Arguments:

  TEMPLATE  nomad job template
    If no argument is given we look for a single *.nomad file

General Options:

  -consul-address=<addr>
    The Consul host and port to use when making Consul KeyValue lookups for
    template rendering.

  -log-level=<level>
    Specify the verbosity level of Levant's logs. Valid values include DEBUG,
    INFO, and WARN, in decreasing order of verbosity. The default is INFO.

  -log-format=<format>
    Specify the format of Levant's logs. Valid values are HUMAN or JSON. The
    default is HUMAN.

  -out=<file>
    Specify the path to write the rendered template out to, if a file exists at
    the specified path it will be truncated before rendering. The template will be
    rendered to stdout if this is not set.

  -var-file=<file>
    The variables file to render the template with. You can repeat this flag multiple
    times to supply multiple var-files. [default: levant.(json|yaml|yml|tf)]
`
	return strings.TrimSpace(helpText)
}

// Synopsis is provides a brief summary of the template command.
func (c *RenderCommand) Synopsis() string {
	return "Render a Nomad job from a template"
}

// Run triggers a run of the Levant template functions.
func (c *RenderCommand) Run(args []string) int {

	var addr, outPath, templateFile string
	var variables []string
	var err error
	var tpl *bytes.Buffer
	var level, format string

	flags := c.Meta.FlagSet("render", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.StringVar(&addr, "consul-address", "", "")
	flags.StringVar(&level, "log-level", "INFO", "")
	flags.StringVar(&format, "log-format", "HUMAN", "")
	flags.Var((*helper.FlagStringSlice)(&variables), "var-file", "")
	flags.StringVar(&outPath, "out", "", "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

	if err = logging.SetupLogger(level, format); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

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

	tpl, err = template.RenderTemplate(templateFile, variables, addr, &c.Meta.flagVars)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	out := os.Stdout
	if outPath != "" {
		out, err = os.Create(outPath)
		if err != nil {
			c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
			return 1
		}
	}

	_, err = tpl.WriteTo(out)
	if err != nil {
		c.UI.Error(fmt.Sprintf("[ERROR] levant/command: %v", err))
		return 1
	}

	return 0
}
