// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"strings"

	"github.com/hashicorp/levant/levant/structs"
	"github.com/hashicorp/levant/logging"
	"github.com/hashicorp/levant/scale"
)

// ScaleInCommand is the command implementation that allows users to scale a
// Nomad job out.
type ScaleInCommand struct {
	Meta
}

// Help provides the help information for the scale-in command.
func (c *ScaleInCommand) Help() string {
	helpText := `
Usage: levant scale-in [options] <job-id>

  Scale a Nomad job and optional task group out.

General Options:

  -address=<http_address>
    The Nomad HTTP API address including port which Levant will use to make
    calls.

  -allow-stale
    Allow stale consistency mode for requests into nomad.
  
  -log-level=<level>
    Specify the verbosity level of Levant's logs. Valid values include DEBUG,
    INFO, and WARN, in decreasing order of verbosity. The default is INFO.
  
  -log-format=<format>
    Specify the format of Levant's logs. Valid values are HUMAN or JSON. The
    default is HUMAN.
	
Scale In Options:

  -count=<num>
    The count by which the job and task groups should be scaled in by. Only
    one of count or percent can be passed.

  -percent=<num>
    A percentage value by which the job and task groups should be scaled in
    by. Counts will be rounded up, to ensure required capacity is met. Only 
    one of count or percent can be passed.

  -task-group=<name>
    The name of the task group you wish to target for scaling. If this is not
    specified, all task groups within the job will be scaled.
`
	return strings.TrimSpace(helpText)
}

// Synopsis is provides a brief summary of the scale-in command.
func (c *ScaleInCommand) Synopsis() string {
	return "Scale in a Nomad job"
}

// Run triggers a run of the Levant scale-in functions.
func (c *ScaleInCommand) Run(args []string) int {

	var err error
	var logL, logF string

	config := &scale.Config{
		Client: &structs.ClientConfig{},
		Scale: &structs.ScaleConfig{
			Direction: structs.ScalingDirectionIn,
		},
	}

	flags := c.Meta.FlagSet("scale-in", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.StringVar(&config.Client.Addr, "address", "", "")
	flags.BoolVar(&config.Client.AllowStale, "allow-stale", false, "")
	flags.StringVar(&logL, "log-level", "INFO", "")
	flags.StringVar(&logF, "log-format", "HUMAN", "")
	flags.IntVar(&config.Scale.Count, "count", 0, "")
	flags.IntVar(&config.Scale.Percent, "percent", 0, "")
	flags.StringVar(&config.Scale.TaskGroup, "task-group", "", "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

	if len(args) != 1 {
		c.UI.Error("This command takes one argument: <job-name>")
		return 1
	}

	config.Scale.JobID = args[0]

	if config.Scale.Count == 0 && config.Scale.Percent == 0 || config.Scale.Count > 0 && config.Scale.Percent > 0 {
		c.UI.Error("You must set either -count or -percent flag to scale-in")
		return 1
	}

	if config.Scale.Count > 0 {
		config.Scale.DirectionType = structs.ScalingDirectionTypeCount
	}

	if config.Scale.Percent > 0 {
		config.Scale.DirectionType = structs.ScalingDirectionTypePercent
	}

	if err = logging.SetupLogger(logL, logF); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	success := scale.TriggerScalingEvent(config)
	if !success {
		return 1
	}

	return 0
}
