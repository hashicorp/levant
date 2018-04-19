package levant

import (
	"errors"
	"sync"
	"time"

	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/jrasell/levant/levant/structs"
	"github.com/jrasell/levant/logging"
)

// StartMonitor will start monitoring a job
func StartMonitor(config *structs.Config, evalID *string, timeoutSeconds uint64, flagVars *map[string]string) error {
	// Create our new deployment object.                                                           |~
	levantDep, err := newLevantDeployment(config)
	if err != nil {
		logging.Error("levant/monitor: unable to setup Levant deployment: %v", err)
		return err
	}

	// start monitoring
	err = levantDep.monitor(evalID, timeoutSeconds)
	if err != nil {
		// we have already reported the error so we don't need to report it again here
		return err
	}
	return nil
}

// Monitor follows a job until it completes.
func (l *levantDeployment) monitor(evalID *string, timeoutSeconds uint64) error {

	// set our timeout
	timeout := time.Tick(time.Second * time.Duration(timeoutSeconds))

	// close this to stop the job monitor
	finish := make(chan bool)
	defer close(finish)
	// set up some communication channels
	jobChan := make(chan *nomad.Job, 1)
	errChan := make(chan error, 1)

	// get the evaluation
	eval, _, err := l.nomad.Evaluations().Info(*evalID, nil)
	if err != nil {
		logging.Error("levant/monitor: unable to get evaluation %v: %v", *evalID, err)
		return err
	}

	// get the job name
	jobName := eval.JobID

	// start our monitor
	go l.monitorJobInfo(jobName, jobChan, errChan, finish)

	for {
		select {
		case <-timeout:
			logging.Error("levant/monitor: timeout reached while monitoring job %s", jobName)
			// run the allocation inspector
			var allocIDS []string
			// get some additional information about the exit of the job
			allocs, _, err := l.nomad.Evaluations().Allocations(*evalID, nil)
			if err != nil {
				logging.Error("levant/monitor: unable to get allocations from evaluation %s: %v", evalID, err)
				return err
			}
			// check to see if any of our allocations failed
			for _, alloc := range allocs {
				for _, task := range alloc.TaskStates {
					// we need to test for success
					if task.State != nomadStructs.TaskStarted {
						allocIDS = append(allocIDS, alloc.ID)
						// once we add the allocation we don't need to add it again
						break
					}
				}
			}

			l.inspectAllocs(allocIDS)
			return errors.New("timeout reached")
		case err = <-errChan:
			logging.Error("levant/monitor: unable to query job %s: %v", jobName, err)
			logging.Error("Retrying...")

		case job := <-jobChan:
			// depending on the state of the job we do different things
			switch *job.Status {
			// if the job is stopped then we take some action depending on why it stopped
			case nomadStructs.JobStatusDead:
				var allocIDS []string
				// get some additional information about the exit of the job
				allocs, _, err := l.nomad.Evaluations().Allocations(*evalID, nil)
				if err != nil {
					logging.Error("levant/monitor: unable to get allocations from evaluation %s: %v", evalID, err)
					return err
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
					l.inspectAllocs(allocIDS)
					return errors.New("Some or all allocations failed")
				}
				// otherwise we print a message and just return no error
				logging.Info("levant/monitor: job %s has status %s", jobName, *job.Status)
				return nil
			case nomadStructs.JobStatusRunning:
				logging.Info("levant/monitor: job %s has status %s", jobName, *job.Status)
				// if its a paramaterized or periodic job we stop here
				if job.IsParameterized() {
					logging.Info("levant/monitor: job %s is parameterized.  Running is its final state.", jobName)
					return nil
				}
				if job.IsPeriodic() {
					logging.Info("levant/monitor: job %s is periodic.  Running is its final state.", jobName)
					return nil
				}
			default:
				logging.Debug("levant/monitor: got job state %s.  Don't know what to do with that.", *job.Status)
			}
		}
	}
}

// monitorJobInfo will get information on a job from nomad and returns the information on channels
// once it has updated
func (l *levantDeployment) monitorJobInfo(jobName string, jobChan chan<- *nomad.Job, errChan chan<- error, done chan bool) {

	// Setup the Nomad QueryOptions to allow blocking query and a timeout.
	q := &nomad.QueryOptions{WaitIndex: 0, WaitTime: time.Second * 10}

	for {
		select {
		case <-done:
			// allow us to exit on demand (technically, it will still need to wait for the namad query to return)
			return
		default:
			// get our job info
			job, meta, err := l.nomad.Jobs().Info(jobName, q)
			if err != nil {
				errChan <- err
				// sleep a bit before retrying
				time.Sleep(time.Second * 5)
				continue
			}
			//logging.Info("lastIndex: %v", meta.LastIndex)
			// only take action if the informaiton has changed
			if meta.LastIndex > q.WaitIndex {
				q.WaitIndex = meta.LastIndex
				jobChan <- job
			} else {
				// log a debug message
				logging.Debug("levant/monitor: job %s currently has status %s", jobName, *job.Status)
			}
		}
	}
}

// inspectAllocs is a helper function that will call the allocInspector for each of the provided allocations
// and return when it has completed
func (l *levantDeployment) inspectAllocs(allocs []string) {
	// if we have failed allocations than we print information about them
	if len(allocs) > 0 {
		// we want to run throuh and get messages for all of our failed allocations in parallel
		var wg sync.WaitGroup
		wg.Add(+len(allocs))

		// Inspect each allocation.
		for _, id := range allocs {
			logging.Debug("levant/monitor: launching allocation inspector for alloc %v", id)
			go l.allocInspector(id, &wg)
		}

		// wait until our allocations have printed messages
		wg.Wait()
		return
	}
}
