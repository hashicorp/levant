// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nomad

import "github.com/hashicorp/nomad/api"

// GenerateBlockingQueryOptions generate Nomad API QueryOptions that can be
// used for blocking. The namespace parameter will be set, if its non-nil.
func GenerateBlockingQueryOptions(ns *string) *api.QueryOptions {
	q := api.QueryOptions{WaitIndex: 1}
	if ns != nil {
		q.Namespace = *ns
	}
	return &q
}
