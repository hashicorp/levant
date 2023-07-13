// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nomad

import (
	"testing"

	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateBlockingQueryOptions(t *testing.T) {
	testCases := []struct {
		inputNS        *string
		expectedOutput *api.QueryOptions
		name           string
	}{
		{
			inputNS: nil,
			expectedOutput: &api.QueryOptions{
				WaitIndex: 1,
			},
			name: "nil input namespace",
		},
		{
			inputNS: stringToPtr("non-default"),
			expectedOutput: &api.QueryOptions{
				WaitIndex: 1,
				Namespace: "non-default",
			},
			name: "non-nil input namespace",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOutput := GenerateBlockingQueryOptions(tc.inputNS)
			assert.Equal(t, tc.expectedOutput, actualOutput, tc.name)
		})
	}
}

func stringToPtr(s string) *string { return &s }
