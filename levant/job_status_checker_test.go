package levant

import (
	"testing"

	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
)

func TestJobStatusChecker_allocationStatusChecker(t *testing.T) {

	// Setup our LevantTask structures.
	levantTasks1 := make(map[string]map[string]string)
	levantTasks1["10246d87-ecd7-21ad-13b2-f0c564647d64"] = make(map[string]string)
	levantTasks1["10246d87-ecd7-21ad-13b2-f0c564647d64"]["task1"] = initialTaskHealth

	levantTasks2 := make(map[string]map[string]string)
	levantTasks2["20246d87-ecd7-21ad-13b2-f0c564647d64"] = make(map[string]string)
	levantTasks2["20246d87-ecd7-21ad-13b2-f0c564647d64"]["task1"] = initialTaskHealth
	levantTasks2["20246d87-ecd7-21ad-13b2-f0c564647d64"]["task2"] = initialTaskHealth

	levantTasks3 := make(map[string]map[string]string)
	levantTasks3["30246d87-ecd7-21ad-13b2-f0c564647d64"] = make(map[string]string)
	levantTasks3["30246d87-ecd7-21ad-13b2-f0c564647d64"]["task1"] = initialTaskHealth

	// Build a small AllocationListStubs with required information.
	var allocs1 []*nomad.AllocationListStub
	taskStates1 := make(map[string]*nomad.TaskState)
	taskStates1["task1"] = &nomad.TaskState{State: nomadStructs.TaskStateRunning}
	allocs1 = append(allocs1, &nomad.AllocationListStub{
		ID:         "10246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates1,
	})

	var allocs2 []*nomad.AllocationListStub
	taskStates2 := make(map[string]*nomad.TaskState)
	taskStates2["task1"] = &nomad.TaskState{State: nomadStructs.TaskStateRunning}
	taskStates2["task2"] = &nomad.TaskState{State: nomadStructs.TaskStatePending}
	allocs2 = append(allocs2, &nomad.AllocationListStub{
		ID:         "20246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates2,
	})

	var allocs3 []*nomad.AllocationListStub
	taskStates3 := make(map[string]*nomad.TaskState)
	taskStates3["task1"] = &nomad.TaskState{State: nomadStructs.TaskStateDead}
	allocs3 = append(allocs3, &nomad.AllocationListStub{
		ID:         "30246d87-ecd7-21ad-13b2-f0c564647d64",
		TaskStates: taskStates3,
	})

	cases := []struct {
		levantTasks    map[string]map[string]string
		allocs         []*nomad.AllocationListStub
		dead           int
		expectedDead   int
		expectedAllocs int
	}{
		{
			levantTasks1,
			allocs1,
			0,
			0,
			0,
		},
		{
			levantTasks2,
			allocs2,
			0,
			0,
			1,
		},
		{
			levantTasks3,
			allocs3,
			0,
			1,
			0,
		},
	}

	for _, tc := range cases {
		allocationStatusChecker(tc.levantTasks, tc.allocs, &tc.dead)

		if len(tc.levantTasks) != tc.expectedAllocs {
			t.Fatalf("expected %v but got %v", tc.expectedAllocs, len(tc.levantTasks))
		}
		if tc.dead != tc.expectedDead {
			t.Fatalf("expected %v dead task(s) but got %v", tc.expectedDead, tc.dead)
		}
	}
}
