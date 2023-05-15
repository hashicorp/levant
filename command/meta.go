// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"bufio"
	"flag"
	"io"

	"github.com/hashicorp/levant/helper"
	"github.com/mitchellh/cli"
)

// FlagSetFlags is an enum to define what flags are present in the
// default FlagSet returned by Meta.FlagSet
type FlagSetFlags uint

// Consts which helps us track meta CLI falgs.
const (
	FlagSetNone        FlagSetFlags = 0
	FlagSetBuildFilter FlagSetFlags = 1 << iota
	FlagSetVars
)

// Meta contains the meta-options and functionality that nearly every
// Levant command inherits.
type Meta struct {
	UI cli.Ui

	// These are set by command-line flags
	flagVars map[string]interface{}
}

// FlagSet returns a FlagSet with the common flags that every
// command implements.
func (m *Meta) FlagSet(n string, fs FlagSetFlags) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)

	// FlagSetVars tells us what variables to use
	if fs&FlagSetVars != 0 {
		f.Var((*helper.Flag)(&m.flagVars), "var", "")
	}

	// Create an io.Writer that writes to our Ui properly for errors.
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
