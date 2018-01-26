package levant

import (
	"sync"
	"time"

	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/jrasell/levant/logging"
)

// checkBatchJob checks that the status of a batch job at least reaches a status of
// running. It can, optionally, monitor a job until it has finished as well.  This
// is required as currently Nomad does not support deployments of job type batch.
// Parameters:
// monitor - whether to follow the task to completion or only to insertion
// timeout - how long to wait until we fail the task
func (c *nomadClient) checkBatchJob(evalID *string, timeoutSeconds uint64, monitor bool) bool {
	// set some timeouts
	timeout := time.Tick(time.Second * time.Duration(timeoutSeconds))
	// close this to stop the job monitor
	finish := make(chan bool)
	defer close(finish)
	// set up some communication channels
	jobChan := make(chan *nomad.Job, 1)
	errChan := make(chan error, 1)

	// get the evaluation
	eval, _, err := c.nomad.Evaluations().Info(*evalID, nil)
	if err != nil {
		logging.Error("levant/deploy: unable to get info of evaluation %s: %v", eval.ID, err)
		return false
	}

	// get the job name
	jobName := eval.JobID

	// start our monitor
	go c.monitorJobInfo(jobName, jobChan, errChan, finish)

	for {
		select {
		case <-timeout:
			logging.Error("levant/job_status_checker: timeout reached while verifying the status of batch job %s", jobName)
			// run the allocation inspector
			var allocIDS []string
			// get some additional information about the exit of the job
			allocs, _, err := c.nomad.Evaluations().Allocations(*evalID, nil)
			if err != nil {
				logging.Error("levant/log_status_checker: unable to get allocations from evaluation %s: %v", evalID, err)
				return false
			}
			// check to see if any of our allocations failed
			for _, alloc := range allocs {
				logging.Info("hello")
				for _, task := range alloc.TaskStates {
					// we need to test for success
					if task.State != nomadStructs.TaskStarted {
						allocIDS = append(allocIDS, alloc.ID)
						// once we add the allocation we don't need to add it again
						break
					}
				}
			}

			c.inspectAllocs(allocIDS)
			return false
		case err = <-errChan:
			logging.Error("levant/job_status_checker: unable to query batch job %s: %v", jobName, err)
			logging.Error("Retrying...")

		case job := <-jobChan:
			// depending on the state of the job we do different things
			switch *job.Status {
			// if the job is stopped then we take some action depending on why it stopped
			case nomadStructs.JobStatusDead:
				var allocIDS []string
				// get some additional information about the exit of the job
				allocs, _, err := c.nomad.Evaluations().Allocations(*evalID, nil)
				if err != nil {
					logging.Error("levant/log_status_checker: unable to get allocations from evaluation %s: %v", evalID, err)
					return false
				}
				// check to see if any of our allocations failed
				for _, alloc := range allocs {
					for _, task := range alloc.TaskStates {
						if task.Failed {
							allocIDS = append(allocIDS, alloc.ID)
						}
					}
				}
				if len(allocIDS) > 0 {
					c.inspectAllocs(allocIDS)
					return false
				} else {
					// otherwise we just return true
					return true
				}
			case nomadStructs.JobStatusRunning:
				logging.Info("levant/job_status_checker: batch job %s has status %s", jobName, *job.Status)
				// if we are not set to wait until the job is done then we exit
				// additionally, regardless of whether we are set to wait, periodic jobs always exit on running
				if !monitor || job.IsPeriodic() {
					return true
				}
			default:
				logging.Debug("levant/job_status_checker: got job state %s.  Don't know what to do with that.", *job.Status)
			}
		}
	}
}

// monitorJobInfo will get information on a job from nomad and returns the information on channels
// once it has updated
func (c *nomadClient) monitorJobInfo(jobName string, jobChan chan<- *nomad.Job, errChan chan<- error, done chan bool) {

	// Setup the Nomad QueryOptions to allow blocking query and a timeout.
	q := &nomad.QueryOptions{WaitIndex: 0, WaitTime: time.Second * 10}

	for {
		select {
		case <-done:
			// allow us to exit on demand (technically, it will still need to wait for the namad query to return)
			return
		default:
			// get our job info
			job, meta, err := c.nomad.Jobs().Info(jobName, q)
			if err != nil {
				errChan <- err
				// sleep a bit before retrying
				time.Sleep(time.Second * 5)
				continue
			}
			logging.Info("lastIndex: %v", meta.LastIndex)
			// only take action if the informaiton has changed
			if meta.LastIndex > q.WaitIndex {
				q.WaitIndex = meta.LastIndex
				jobChan <- job
			} else {
				// log a debug message
				logging.Debug("levant/job_status_checker: batch job %s currently has status %s", jobName, *job.Status)
			}
		}
	}
}

// inspectAllocs is a helper function that will call the allocInspector for each of the provided allocations
// and return when it has completed
func (c *nomadClient) inspectAllocs(allocs []string) {
	// if we have failed allocations than we print information about them
	if len(allocs) > 0 {
		// we want to run throuh and get messages for all of our failed allocations in parallel
		var wg sync.WaitGroup
		wg.Add(+len(allocs))

		// Inspect each allocation.
		for _, id := range allocs {
			logging.Debug("levant/failure_inspector: launching allocation inspector for alloc %v", id)
			go c.allocInspector(id, &wg)
		}

		// wait until our allocations have printed messages
		wg.Wait()
		return
	}
}
