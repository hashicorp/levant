// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package test

import (
	"testing"

	"github.com/hashicorp/levant/test/acctest"
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

func TestDeploy_basicJson(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_basic.nomad.json",
					IsJSON:      true,
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_invalidJson(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_basic_invalid.nomad.json",
					IsJSON:      true,
				},
				ExpectErr: true,
				CheckErr: func(err error) bool {
					return err.Error() == "error rendering template: Failed to parse JSON job: json: cannot unmarshal array into Go struct field Job.Job.Type of type string"
				},
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_jsonWithoutIdName(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_jsonWithoutIdName.nomad.json",
					IsJSON:      true,
				},
				ExpectErr: true,
				CheckErr: func(err error) bool {
					return err.Error() == "error rendering template: JSON is missing ID field"
				},
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_notJson(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_basic.nomad",
					IsJSON:      true,
				},
				ExpectErr: true,
				CheckErr: func(err error) bool {
					return err.Error() == "error rendering template: Detected JSON but failed to parse JSON job: invalid character '#' looking for beginning of value"
				},
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_driverError(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_driver_error.nomad",
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

func TestDeploy_allocError(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_alloc_error.nomad",
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

func TestDeploy_count(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_count.nomad",
					Vars: map[string]interface{}{
						"count": "3",
					},
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_count.nomad",
					Vars: map[string]interface{}{
						"count": "1",
					},
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
			{
				// expect levant to read counts from the api
				Check: acctest.CheckTaskGroupCount("test", 3),
			},
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_count.nomad",
					Vars: map[string]interface{}{
						"count": "1",
					},
					ForceCount: true,
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
			{
				Check: acctest.CheckTaskGroupCount("test", 1),
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_canary(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_canary.nomad",
					Canary:      10,
					Vars: map[string]interface{}{
						"env_version": "1",
					},
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_canary.nomad",
					Canary:      10,
					Vars: map[string]interface{}{
						"env_version": "2",
					},
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_lifecycle(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_lifecycle.nomad",
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}

func TestDeploy_taskScalingStanza(t *testing.T) {
	acctest.Test(t, acctest.TestCase{
		Steps: []acctest.TestStep{
			{
				Runner: acctest.DeployTestStepRunner{
					FixtureName: "deploy_task_scaling.nomad",
				},
				Check: acctest.CheckDeploymentStatus("successful"),
			},
		},
		CleanupFunc: acctest.CleanupPurgeJob,
	})
}
