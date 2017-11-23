package command

import (
	"testing"

	"github.com/jrasell/levant/levant"
)

func TestDeploy_checkCanaryAutoPromote(t *testing.T) {

	fVars := make(map[string]string)
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
		job, err := levant.RenderJob(c.File, "", &fVars)
		if err != nil {
			t.Fatalf("case %d failed: %v", i, err)
		}

		out := depCommand.checkCanaryAutoPromote(job, c.CanaryPromote)
		if out != c.Output {
			t.Fatalf("case %d: got \"%v\"; want %v", i, out, c.Output)
		}
	}
}
