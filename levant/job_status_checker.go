package levant

import (
	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/jrasell/levant/logging"
)

// initialTaskHealth is the Levant health status assosiated to a Task when it is
// initially discovered as part of the deployment.
const initialTaskHealth = "unknown"

// jobStatusChecker checks the status of a job at least reaches a status of
// running. Depending on the type of job and its configuration it can go through
// more checks.
func (l *levantDeployment) jobStatusChecker(evalID *string) bool {

	logging.Debug("levant/job_status_checker: running job status checker for job %s",
		*l.config.Job.Name)

	// Run the initial job status check to ensure the job reaches a state of
	// running.
	jStatus := l.simpleJobStatusChecker(*l.config.Job.ID)

	// Periodic and parameterized batch jobs do not produce evaluations and so
	// can only go through the simplest of checks.
	if *evalID == "" {
		return jStatus
	}

	// Job registrations that produce an evaluation can be more thoroughly
	// checked even if they don't support Nomad deployments.
	if jStatus {
		return l.jobAllocationChecker(evalID)
	}

	return false
}

// simpleJobStatusChecker is used to check that jobs which do not emit initial
// evaluations at least reach a job status of running.
func (l *levantDeployment) simpleJobStatusChecker(jobID string) bool {

	j := l.config.Job.Name
	q := &nomad.QueryOptions{WaitIndex: 1}

	for {

		job, meta, err := l.nomad.Jobs().Info(*j, q)
		if err != nil {
			logging.Error("levant/job_status_checker: unable to query job %s: %v", *j, err)
			return false
		}

		// If the LastIndex is not greater than our stored LastChangeIndex, we don't
		// need to do anything.
		if meta.LastIndex <= q.WaitIndex {
			continue
		}

		// Checks the status of the job and proceed as expected depending on this.
		switch *job.Status {
		case nomadStructs.JobStatusRunning:
			logging.Info("levant/job_status_checker: job %s has status %s", *j, *job.Status)
			return true
		case nomadStructs.JobStatusPending:
			q.WaitIndex = meta.LastIndex
			continue
		case nomadStructs.JobStatusDead:
			logging.Error("levant/job_status_checker: job %s has status %s", *j, *job.Status)
			return false
		}
	}
}

// jobAllocationChecker is the main entry point into the allocation checker for
// jobs that do not support Nomad deployments.
func (l *levantDeployment) jobAllocationChecker(evalID *string) bool {

	// Track if we experience any dead tasks in a dumb way.
	var deadTasks int

	j := l.config.Job.Name
	q := &nomad.QueryOptions{WaitIndex: 1}

	// Build our small internal checking struct.
	levantTasks := l.buildAllocationChecker(evalID)

	for {

		allocs, meta, err := l.nomad.Evaluations().Allocations(*evalID, q)
		if err != nil {
			logging.Error("levant/job_status_checker: unable to query allocs of job %s: %v",
				*j, err)
			return false
		}

		// If the LastIndex is not greater than our stored LastChangeIndex, we don't
		// need to do anything.
		if meta.LastIndex <= q.WaitIndex {
			continue
		}

		// If we get here, set the wi to the latest Index.
		q.WaitIndex = meta.LastIndex

		allocationStatusChecker(levantTasks, allocs, &deadTasks)

		// If we have no allocations left to track then we can exit and log
		// information depending on the success.
		if len(levantTasks) == 0 && deadTasks == 0 {
			logging.Info("levant/job_status_checker: all allocations in deployment of job %s are running", *j)
			return true
		} else if len(levantTasks) == 0 && deadTasks > 0 {
			return false
		}
	}
}

func (l *levantDeployment) buildAllocationChecker(evalID *string) map[string]map[string]string {

	// Create our map to track allocations and tasks within the evaluation created
	// by the job registration.
	levantTasks := make(map[string]map[string]string)

	q := &nomad.QueryOptions{WaitIndex: 1}

	// We use a for loop here as, during testing, I have observed Levant runs
	// faster than Nomad and so without a blocking query a pure GET request can
	// trigger before Nomad builds the allocation.
	for {

		// Pull the latest information abou the evaluation allocations from Nomad.
		allocs, meta, err := l.nomad.Evaluations().Allocations(*evalID, q)
		if err != nil {
			logging.Error("levant/job_status_checker: unable to query evaluation for allocations: %v", err)
			return nil
		}

		// If the LastIndex is not greater than our stored LastChangeIndex, we don't
		// need to do anything.
		if meta.LastIndex <= q.WaitIndex {
			continue
		}

		// If we get here, set the wi to the latest Index.
		q.WaitIndex = meta.LastIndex

		// Iterate over each allocation which can contain multiple tasks in order to
		// build our object of tasks to check.
		for _, alloc := range allocs {
			for taskName := range alloc.TaskStates {

				// If we have not seen the allocation previously we need to init the map.
				if levantTasks[alloc.ID] == nil {
					levantTasks[alloc.ID] = make(map[string]string)
				}

				// Set the task health to our initial status of Unknown.
				levantTasks[alloc.ID][taskName] = initialTaskHealth
			}
		}

		if len(levantTasks) == 0 {
			continue
		}

		return levantTasks
	}
}

// allocationStatusChecker is used to check the state of allocations within a
// job deployment, an update Levants internal tracking on task status based on
// this. This functionality exists as Nomad does not currently support
// deployments across all job types.
func allocationStatusChecker(levantTasks map[string]map[string]string, allocs []*nomad.AllocationListStub, deadTask *int) {

	for _, alloc := range allocs {
		for taskName, task := range alloc.TaskStates {
			levantTasks[alloc.ID][taskName] = task.State

			// If the task is running, remove it from tracking.
			switch levantTasks[alloc.ID][taskName] {
			case nomadStructs.TaskStateRunning:
				logging.Info("levant/job_status_checker: task %s in allocation %s has reached %s state",
					taskName, alloc.ID, nomadStructs.TaskStateRunning)
				delete(levantTasks[alloc.ID], taskName)

			case nomadStructs.TaskStatePending:
				logging.Debug("levant/job_status_checker: task %s in allocation %s now in running state",
					taskName, alloc.ID)

			// If the task is dead, incrament the deadTask counter and remove the task
			// from tracking.
			case nomadStructs.TaskStateDead:
				logging.Error("levant/job_status_checker: task %s in allocation %s now in dead state",
					taskName, alloc.ID)
				*deadTask++
				delete(levantTasks[alloc.ID], taskName)
			}

			// If we have no tasks left under the allocation to track remove the
			// allocation from our tracker.
			if len(levantTasks[alloc.ID]) == 0 {
				delete(levantTasks, alloc.ID)
			}
		}
	}
}
