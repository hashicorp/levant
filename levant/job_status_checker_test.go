// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package levant

import (
	"testing"

	nomad "github.com/hashicorp/nomad/api"
)

func TestJobStatusChecker_allocationStatusChecker(t *testing.T) {

	// Build our task status maps
	levantTasks1 := make(map[TaskCoordinate]string)
	levantTasks2 := make(map[TaskCoordinate]string)
	levantTasks3 := make(map[TaskCoordinate]string)

	// Build a small AllocationListStubs with required information.
	var allocs1 []*nomad.AllocationListStub
	taskStates1 := make(map[string]*nomad.TaskState)
	taskStates1["task1"] = &nomad.TaskState{State: "running"}
	allocs1 = append(allocs1, &nomad.AllocationListStub{
		ID:         "10246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates1,
	})

	var allocs2 []*nomad.AllocationListStub
	taskStates2 := make(map[string]*nomad.TaskState)
	taskStates2["task1"] = &nomad.TaskState{State: "running"}
	taskStates2["task2"] = &nomad.TaskState{State: "pending"}
	allocs2 = append(allocs2, &nomad.AllocationListStub{
		ID:         "20246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates2,
	})

	var allocs3 []*nomad.AllocationListStub
	taskStates3 := make(map[string]*nomad.TaskState)
	taskStates3["task1"] = &nomad.TaskState{State: "dead"}
	allocs3 = append(allocs3, &nomad.AllocationListStub{
		ID:         "30246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates3,
	})

	cases := []struct {
		levantTasks      map[TaskCoordinate]string
		allocs           []*nomad.AllocationListStub
		dead             int
		expectedDead     int
		expectedComplete bool
	}{
		{
			levantTasks1,
			allocs1,
			0,
			0,
			true,
		},
		{
			levantTasks2,
			allocs2,
			0,
			0,
			false,
		},
		{
			levantTasks3,
			allocs3,
			0,
			1,
			true,
		},
	}

	for _, tc := range cases {
		complete, dead := allocationStatusChecker(tc.levantTasks, tc.allocs)

		if complete != tc.expectedComplete {
			t.Fatalf("expected complete to be %v but got %v", tc.expectedComplete, complete)
		}
		if dead != tc.expectedDead {
			t.Fatalf("expected %v dead task(s) but got %v", tc.expectedDead, dead)
		}
	}
}
