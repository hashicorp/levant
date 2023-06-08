// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"github.com/mitchellh/cli"
)

var _ cli.Command = &VersionCommand{}

// VersionCommand is a Command implementation that prints the version.
type VersionCommand struct {
	Version string
	UI      cli.Ui
}

// Help provides the help information for the version command.
func (c *VersionCommand) Help() string {
	return ""
}

// Synopsis is provides a brief summary of the version command.
func (c *VersionCommand) Synopsis() string {
	return "Prints the Levant version"
}

// Run executes the version command.
func (c *VersionCommand) Run(_ []string) int {
	c.UI.Info(c.Version)
	return 0
}
