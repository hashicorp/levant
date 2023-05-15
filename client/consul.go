// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	consul "github.com/hashicorp/consul/api"
)

// NewConsulClient is used to create a new client to interact with Consul.
func NewConsulClient(addr string) (*consul.Client, error) {
	config := consul.DefaultConfig()

	if addr != "" {
		config.Address = addr
	}

	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	return c, nil
}
