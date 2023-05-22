// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	nomad "github.com/hashicorp/nomad/api"
)

// TestCase is a single test of levant
type TestCase struct {
	// Steps are ran in order stopping on failure
	Steps []TestStep

	// SetupFunc is called at before the steps
	SetupFunc TestStateFunc

	// CleanupFunc is called at the end of the TestCase
	CleanupFunc TestStateFunc
}

// TestStep is a single step within a TestCase
type TestStep struct {
	// Runner is used to execute the step, can be nil for a check only step
	Runner TestStepRunner

	// Check is called after Runner if it does not fail
	Check TestStateFunc

	// ExpectErr allows Runner to fail, use CheckErr to confirm error is correct
	ExpectErr bool

	// CheckErr is called if Runner fails and ExpectErr is true
	CheckErr func(error) bool
}

// TestStepRunner models a runner for a TestStep
type TestStepRunner interface {
	// Run executes the levant feature under testing
	Run(*TestState) error
}

// TestStateFunc is used to verify acceptance test criteria
type TestStateFunc func(*TestState) error

// TestState is the configuration for the TestCase
type TestState struct {
	JobName   string
	Namespace string
	Nomad     *nomad.Client
}

// Test executes a single TestCase
func Test(t *testing.T, c TestCase) {
	if len(c.Steps) < 1 {
		t.Fatal("must have at least one test step")
	}

	nomad, err := newNomadClient()
	if err != nil {
		t.Fatalf("failed to create nomad client: %s", err)
	}

	state := &TestState{
		JobName: fmt.Sprintf("levant-%s", t.Name()),
		Nomad:   nomad,
	}

	if c.CleanupFunc != nil {
		t.Cleanup(func() {
			err = c.CleanupFunc(state)
			if err != nil {
				t.Errorf("cleanup failed: %s", err)
			}
		})
	}

	if c.SetupFunc != nil {
		if err := c.SetupFunc(state); err != nil {
			t.Errorf("setup failed: %s", err)
		}
	}

	for i, step := range c.Steps {
		stepNum := i + 1

		if step.Runner != nil {
			err = step.Runner.Run(state)
			if err != nil {
				if !step.ExpectErr {
					t.Errorf("step %d/%d failed: %s", stepNum, len(c.Steps), err)
					break
				}

				if step.CheckErr != nil {
					ok := step.CheckErr(err)
					if !ok {
						t.Errorf("step %d/%d CheckErr failed: %s", stepNum, len(c.Steps), err)
						break
					}
				}
			}
		}

		if step.Check != nil {
			err = step.Check(state)
			if err != nil {
				t.Errorf("step %d/%d Check failed: %s", stepNum, len(c.Steps), err)
				break
			}
		}
	}

}

// CleanupPurgeJob is a cleanup func to purge the TestCase job from Nomad
func CleanupPurgeJob(s *TestState) error {
	_, _, err := s.Nomad.Jobs().Deregister(s.JobName, true, &nomad.WriteOptions{Namespace: s.Namespace})
	return err
}

// CleanupPurgeJobAndNamespace is a cleanup func to purge the TestCase job and
// test namespace from Nomad
func CleanupPurgeJobAndNamespace(s *TestState) error {
	var mErr error
	if _, _, err := s.Nomad.Jobs().Deregister(s.JobName, true, &nomad.WriteOptions{Namespace: s.Namespace}); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	if _, err := s.Nomad.Namespaces().Delete(s.Namespace, nil); err != nil {
		mErr = multierror.Append(mErr, err)
	}
	return mErr
}

func SetupTestNamespace(ns string) func(s *TestState) error {
	return func(s *TestState) error {
		_, err := s.Nomad.Namespaces().Register(
			&nomad.Namespace{
				Name:        ns,
				Description: "Levant acceptance testing namespace",
			}, nil)
		if err != nil {
			return err
		}
		s.Namespace = ns
		return nil
	}
}

// CheckDeploymentStatus is a TestStateFunc to check if the latest deployment of
// the TestCase job matches the desired status
func CheckDeploymentStatus(status string) TestStateFunc {
	return func(s *TestState) error {
		deploy, _, err := s.Nomad.Jobs().LatestDeployment(s.JobName, &nomad.QueryOptions{Namespace: s.Namespace})
		if err != nil {
			return err
		}

		if deploy == nil {
			return fmt.Errorf("no deployment found for job %s", s.JobName)
		}

		if deploy.Status != status {
			return fmt.Errorf("deployment %s is in status '%s', expected '%s'", deploy.ID, deploy.Status, status)
		}

		return nil
	}
}

// CheckTaskGroupCount is a TestStateFunc to check a TaskGroup count
func CheckTaskGroupCount(groupName string, count int) TestStateFunc {
	return func(s *TestState) error {
		job, _, err := s.Nomad.Jobs().Info(s.JobName, nil)
		if err != nil {
			return err
		}

		for _, group := range job.TaskGroups {
			if groupName == *group.Name {
				if *group.Count == count {
					return nil
				}

				return fmt.Errorf("task group %s count is %d, expected %d", groupName, *group.Count, count)
			}
		}

		return fmt.Errorf("unable to find task group %s", groupName)
	}
}

// newNomadClient creates a Nomad API client configurable by NOMAD_
// env variables or returns an error if Nomad is in an unhealthy state
func newNomadClient() (*nomad.Client, error) {
	c, err := nomad.NewClient(nomad.DefaultConfig())
	if err != nil {
		return nil, err
	}

	resp, err := c.Agent().Health()
	if err != nil {
		return nil, err
	}

	if (resp.Server != nil && !resp.Server.Ok) || (resp.Client != nil && !resp.Client.Ok) {
		return nil, fmt.Errorf("agent unhealthy")
	}
	return c, nil
}
