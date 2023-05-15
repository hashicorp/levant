// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package levant

import (
	nomadHelper "github.com/hashicorp/levant/helper/nomad"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog/log"
)

// TaskCoordinate is a coordinate for an allocation/task combination
type TaskCoordinate struct {
	Alloc    string
	TaskName string
}

// jobStatusChecker checks the status of a job at least reaches a status of
// running. Depending on the type of job and its configuration it can go through
// more checks.
func (l *levantDeployment) jobStatusChecker(evalID *string) bool {

	log.Debug().Msgf("levant/job_status_checker: running job status checker for job")

	// Run the initial job status check to ensure the job reaches a state of
	// running.
	jStatus := l.simpleJobStatusChecker()

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
func (l *levantDeployment) simpleJobStatusChecker() bool {

	q := nomadHelper.GenerateBlockingQueryOptions(l.config.Template.Job.Namespace)

	for {

		job, meta, err := l.nomad.Jobs().Info(*l.config.Template.Job.Name, q)
		if err != nil {
			log.Error().Err(err).Msg("levant/job_status_checker: unable to query job information from Nomad")
			return false
		}

		// If the LastIndex is not greater than our stored LastChangeIndex, we don't
		// need to do anything.
		if meta.LastIndex <= q.WaitIndex {
			continue
		}

		// Checks the status of the job and proceed as expected depending on this.
		switch *job.Status {
		case "running":
			log.Info().Msgf("levant/job_status_checker: job has status %s", *job.Status)
			return true
		case "pending":
			log.Debug().Msgf("levant/job_status_checker: job has status %s", *job.Status)
			q.WaitIndex = meta.LastIndex
			continue
		case "dead":
			log.Error().Msgf("levant/job_status_checker: job has status %s", *job.Status)
			return false
		}
	}
}

// jobAllocationChecker is the main entry point into the allocation checker for
// jobs that do not support Nomad deployments.
func (l *levantDeployment) jobAllocationChecker(evalID *string) bool {

	q := nomadHelper.GenerateBlockingQueryOptions(l.config.Template.Job.Namespace)

	// Build our small internal checking struct.
	levantTasks := make(map[TaskCoordinate]string)

	for {

		allocs, meta, err := l.nomad.Evaluations().Allocations(*evalID, q)
		if err != nil {
			log.Error().Err(err).Msg("levant/job_status_checker: unable to query allocs of job from Nomad")
			return false
		}

		// If the LastIndex is not greater than our stored LastChangeIndex, we don't
		// need to do anything.
		if meta.LastIndex <= q.WaitIndex {
			continue
		}

		// If we get here, set the wi to the latest Index.
		q.WaitIndex = meta.LastIndex

		complete, deadTasks := allocationStatusChecker(levantTasks, allocs)

		// depending on how we finished up we report our status
		// If we have no allocations left to track then we can exit and log
		// information depending on the success.
		if complete && deadTasks == 0 {
			log.Info().Msg("levant/job_status_checker: all allocations in deployment of job are running")
			return true
		} else if complete && deadTasks > 0 {
			return false
		}
	}
}

// allocationStatusChecker is used to check the state of allocations within a
// job deployment, an update Levants internal tracking on task status based on
// this. This functionality exists as Nomad does not currently support
// deployments across all job types.
func allocationStatusChecker(levantTasks map[TaskCoordinate]string, allocs []*nomad.AllocationListStub) (bool, int) {

	complete := true
	deadTasks := 0

	for _, alloc := range allocs {
		for taskName, task := range alloc.TaskStates {
			// if the state is one we haven't seen yet then we print a message
			if levantTasks[TaskCoordinate{alloc.ID, taskName}] != task.State {
				log.Info().Msgf("levant/job_status_checker: task %s in allocation %s now in %s state",
					taskName, alloc.ID, task.State)
				// then we record the new state
				levantTasks[TaskCoordinate{alloc.ID, taskName}] = task.State
			}

			// then we have some case specific actions
			switch levantTasks[TaskCoordinate{alloc.ID, taskName}] {
			// if a task is still pendign we are not yet done
			case "pending":
				complete = false
				// if the task is dead we record that
			case "dead":
				deadTasks++
			}
		}
	}
	return complete, deadTasks
}
