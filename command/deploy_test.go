// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"testing"

	"github.com/hashicorp/levant/template"
)

func TestDeploy_checkCanaryAutoPromote(t *testing.T) {

	fVars := make(map[string]interface{})
	depCommand := &DeployCommand{}
	canaryPromote := 30

	cases := []struct {
		File          string
		CanaryPromote int
		Output        error
	}{
		{
			File:          "test-fixtures/job_canary.nomad",
			CanaryPromote: canaryPromote,
			Output:        nil,
		},
		{
			File:          "test-fixtures/group_canary.nomad",
			CanaryPromote: canaryPromote,
			Output:        nil,
		},
	}

	for i, c := range cases {
		job, err := template.RenderJob(c.File, []string{}, "", &fVars)
		if err != nil {
			t.Fatalf("case %d failed: %v", i, err)
		}

		out := depCommand.checkCanaryAutoPromote(job, c.CanaryPromote)
		if out != c.Output {
			t.Fatalf("case %d: got \"%v\"; want %v", i, out, c.Output)
		}
	}
}

func TestDeploy_checkForceBatch(t *testing.T) {

	fVars := make(map[string]interface{})
	depCommand := &DeployCommand{}
	forceBatch := true

	cases := []struct {
		File       string
		ForceBatch bool
		Output     error
	}{
		{
			File:       "test-fixtures/periodic_batch.nomad",
			ForceBatch: forceBatch,
			Output:     nil,
		},
	}

	for i, c := range cases {
		job, err := template.RenderJob(c.File, []string{}, "", &fVars)
		if err != nil {
			t.Fatalf("case %d failed: %v", i, err)
		}

		out := depCommand.checkForceBatch(job, c.ForceBatch)
		if out != c.Output {
			t.Fatalf("case %d: got \"%v\"; want %v", i, out, c.Output)
		}
	}
}
