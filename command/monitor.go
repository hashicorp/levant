package command

import (
	"strings"

	"github.com/jrasell/levant/levant"
	"github.com/jrasell/levant/levant/structs"
	"github.com/jrasell/levant/logging"
)

// MonitorCommand is the command implementation that allows users to monitor
// a job until it completes
type MonitorCommand struct {
	args []string
	Meta
}

// Help provides the help information for the template command.
func (c *MonitorCommand) Help() string {
	helpText := `
Usage: levant monitor [options] [EVAL_ID]

  Monitor a Nomad job until it completes

Arguments:

  EVAL_ID nomad job evaluation

General Options:
	
  -timeout=<int>
    Number of seconds to allow until we exit with an error.

  -log-level=<level>
    Specify the verbosity level of Levant's logs. Valid values include DEBUG,
    INFO, and WARN, in decreasing order of verbosity. The default is INFO.
`
	return strings.TrimSpace(helpText)
}

// Synopsis is provides a brief summary of the template command.
func (c *MonitorCommand) Synopsis() string {
	return "Monitor a Nomad job until it completes"
}

// Run triggers a run of the Levant monitor function.
func (c *MonitorCommand) Run(args []string) int {

	var timeout uint64
	var evalID string
	var err error
	config := &structs.Config{}

	flags := c.Meta.FlagSet("monitor", FlagSetVars)
	flags.Usage = func() { c.UI.Output(c.Help()) }

	flags.Uint64Var(&timeout, "timeout", 0, "")
	flags.StringVar(&config.LogLevel, "log-level", "INFO", "")

	if err = flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()

	logging.SetLevel(config.LogLevel)

	if len(args) == 1 {
		evalID = args[0]
	} else if len(args) == 0 {
		c.UI.Error(c.Help())
		c.UI.Error("\nERROR: Please provide an evaluation ID to monitor.")
		return 1
	} else {
		c.UI.Error(c.Help())
		return 1
	}

	// trigger our monitor
	err = levant.StartMonitor(config, &evalID, timeout, &c.Meta.flagVars)
	if err != nil {
		// we have already reported the errors so we don't need to do it again here.
		return 1
	}

	return 0
}
