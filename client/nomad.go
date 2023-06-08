// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	nomad "github.com/hashicorp/nomad/api"
)

// NewNomadClient is used to create a new client to interact with Nomad.
func NewNomadClient(addr string) (*nomad.Client, error) {
	config := nomad.DefaultConfig()

	if addr != "" {
		config.Address = addr
	}

	c, err := nomad.NewClient(config)
	if err != nil {
		return nil, err
	}

	return c, nil
}
