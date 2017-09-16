package command

import (
	"bufio"
	"flag"
	"io"

	"github.com/jrasell/levant/helper"
	"github.com/mitchellh/cli"
)

// FlagSetFlags is an enum to define what flags are present in the
// default FlagSet returned by Meta.FlagSet
type FlagSetFlags uint

const (
	FlagSetNone        FlagSetFlags = 0
	FlagSetBuildFilter FlagSetFlags = 1 << iota
	FlagSetVars
)

// Meta contains the meta-options and functionality that nearly every
// Packer command inherits.
type Meta struct {
	UI cli.Ui

	// These are set by command-line flags
	flagVars map[string]string
}

// FlagSet returns a FlagSet with the common flags that every
// command implements. The exact behavior of FlagSet can be configured
// using the flags as the second parameter, for example to disable
// build settings on the commands that don't handle builds.
func (m *Meta) FlagSet(n string, fs FlagSetFlags) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

	// FlagSetVars tells us what variables to use
	if fs&FlagSetVars != 0 {
		f.Var((*helper.Flag)(&m.flagVars), "var", "")
	}

	// Create an io.Writer that writes to our Ui properly for errors.
	// This is kind of a hack, but it does the job. Basically: create
	// a pipe, use a scanner to break it into lines, and output each line
	// to the UI. Do this forever.
	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	go func() {
		for errScanner.Scan() {
			m.UI.Error(errScanner.Text())
		}
	}()
	f.SetOutput(errW)

	return f
}
