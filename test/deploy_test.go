package test

import (
	"testing"

	"github.com/jrasell/levant/test/acctest"
)

func TestDeploy_basic(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_basic.nomad",
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_failure(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_failure.nomad",
				},
				ExpectErr: true,
				CheckErr: func(err error) bool {
					// this is a bit pointless without the error bubbled up from levant
					return true
				},
			},
			{
				// allows us to check a job was registered and previous step error wasn't a parse failure etc.
				Check: acctest.CheckDeploymentStatus("failed"),
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}
