package levant

import (
	"testing"

	nomad "github.com/hashicorp/nomad/api"
)

func TestJobStatusChecker_allocationStatusChecker(t *testing.T) {

	// Build our task status maps
	levantTasks1 := make(map[TaskCoordinate]*nomad.TaskState)
	levantTasks2 := make(map[TaskCoordinate]*nomad.TaskState)
	levantTasks3 := make(map[TaskCoordinate]*nomad.TaskState)
	levantTasks4 := make(map[TaskCoordinate]*nomad.TaskState)

	// Build a small AllocationListStubs with required information.
	var allocs1 []*nomad.AllocationListStub
	taskStates1 := make(map[string]*nomad.TaskState)
	taskStates1["task1"] = &nomad.TaskState{State: "running", Failed: false}
	allocs1 = append(allocs1, &nomad.AllocationListStub{
		ID:         "10246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates1,
	})

	var allocs2 []*nomad.AllocationListStub
	taskStates2 := make(map[string]*nomad.TaskState)
	taskStates2["task1"] = &nomad.TaskState{State: "running", Failed: false}
	taskStates2["task2"] = &nomad.TaskState{State: "pending", Failed: false}
	taskStates2["task3"] = &nomad.TaskState{State: "dead", Failed: false}
	allocs2 = append(allocs2, &nomad.AllocationListStub{
		ID:         "20246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates2,
	})

	var allocs3 []*nomad.AllocationListStub
	taskStates3 := make(map[string]*nomad.TaskState)
	taskStates3["task1"] = &nomad.TaskState{State: "dead", Failed: false}
	allocs3 = append(allocs3, &nomad.AllocationListStub{
		ID:         "30246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates3,
	})

	var allocs4 []*nomad.AllocationListStub
	taskStates4 := make(map[string]*nomad.TaskState)
	taskStates4["task1"] = &nomad.TaskState{State: "dead", Failed: false}
	taskStates4["task2"] = &nomad.TaskState{State: "dead", Failed: true}
	allocs4 = append(allocs4, &nomad.AllocationListStub{
		ID:         "40246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates4,
	})

	cases := []struct {
		levantTasks      map[TaskCoordinate]*nomad.TaskState
		allocs           []*nomad.AllocationListStub
		expectedFailed   int
		expectedComplete bool
	}{
		{
			levantTasks1,
			allocs1,
			0,
			true,
		},
		{
			levantTasks2,
			allocs2,
			0,
			false,
		},
		{
			levantTasks3,
			allocs3,
			0,
			true,
		},
		{
			levantTasks4,
			allocs4,
			1,
			true,
		},
	}

	for _, tc := range cases {
		complete, failed := allocationStatusChecker(tc.levantTasks, tc.allocs)

		if complete != tc.expectedComplete {
			t.Fatalf("expected complete to be %v but got %v", tc.expectedComplete, complete)
		}
		if failed != tc.expectedFailed {
			t.Fatalf("expected %v failed task(s) but got %v", tc.expectedFailed, failed)
		}
	}
}
