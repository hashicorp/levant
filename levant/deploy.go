package levant

import (
	"fmt"
	"strings"
	"time"

	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/jrasell/levant/logging"
)

type nomadClient struct {
	nomad *nomad.Client
}

// NomadClient is an interface to the Nomad API and deployment functions.
type NomadClient interface {
	// Deploy triggers a register of the job resulting in a Nomad deployment which
	// is monitored to determine the eventual state.
	Deploy(*nomad.Job, int, bool) bool
}

// NewNomadClient is used to create a new client to interact with Nomad.
func NewNomadClient(addr string) (NomadClient, error) {
	config := nomad.DefaultConfig()

	if addr != "" {
		config.Address = addr
	}

	c, err := nomad.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &nomadClient{nomad: c}, nil
}

// Deploy triggers a register of the job resulting in a Nomad deployment which
// is monitored to determine the eventual state.
func (c *nomadClient) Deploy(job *nomad.Job, autoPromote int, forceCount bool) (success bool) {

	// Validate the job to check it is syntactically correct.
	if _, _, err := c.nomad.Jobs().Validate(job, nil); err != nil {
		logging.Error("levant/deploy: job validation failed: %v", err)
		return
	}

	// If job.Type isn't set we can't continue
	if job.Type == nil {
		logging.Error("levant/deploy: Nomad job `type` is not set; should be set to `%s`, `%s` or `%s`",
			nomadStructs.JobTypeBatch, nomadStructs.JobTypeSystem, nomadStructs.JobTypeService)
		return
	}

	if !forceCount {
		logging.Debug("levant/deploy: running dynamic job count updater for job %s", *job.Name)
		if err := c.dynamicGroupCountUpdater(job); err != nil {
			return
		}
	}

	// Check that the job has at least 1 TaskGroup with count > 0 (GH-16) if the
	// job is not a System job. Systems jobs do not define counts so cannot be
	// checked.
	if *job.Type != nomadStructs.JobTypeSystem {
		tgCount := 0
		for _, group := range job.TaskGroups {
			tgCount += *group.Count
		}
		if tgCount == 0 {
			logging.Error("levant/deploy: all TaskGroups have a count of 0, nothing to do")
			return
		}
	}

	logging.Info("levant/deploy: triggering a deployment of job %s", *job.Name)

	eval, _, err := c.nomad.Jobs().Register(job, nil)
	if err != nil {
		logging.Error("levant/deploy: unable to register job %s with Nomad: %v", *job.Name, err)
		return
	}

	// Periodic and parameterized jobs do not return an evaluation and therefore
	// can't perform the evaluationInspector.
	if !job.IsPeriodic() && !job.IsParameterized() {

		// Trigger the evaluationInspector to identify any potential errors in the
		// Nomad evaluation run. As far as I can tell from testing; a single alloc
		// failure in an evaluation means no allocs will be placed so we exit here.
		err = c.evaluationInspector(&eval.EvalID)
		if err != nil {
			logging.Error("levant/deploy: %v", err)
			return
		}
	}

	switch *job.Type {
	case nomadStructs.JobTypeService:
		logging.Info("levant/deploy: beginning deployment watcher for job %s", *job.Name)

		// Get the deploymentID from the evaluationID so that we can watch the
		// deployment for end status.
		depID, err := c.getDeploymentID(eval.EvalID)
		if err != nil {
			logging.Error("levant/deploy: unable to get info of evaluation %s: %v", eval.EvalID, err)
			return
		}

		// Get the success of the deployment.
		success = c.deploymentWatcher(depID, autoPromote)

		// If the deployment has not been successful; check whether the job is
		// configured to auto-revert so that this can be tracked.
		if !success {
			dep, _, err := c.nomad.Deployments().Info(depID, nil)
			if err != nil {
				logging.Error("levant/deploy: unable to query deployment %s for auto-revert check: %v",
					dep.ID, err)
				break
			}
			c.checkAutoRevert(dep)
		}

	case nomadStructs.JobTypeBatch:
		return c.checkBatchJob(job.Name)

	default:
		logging.Debug("levant/deploy: job type %s does not support Nomad deployment model", *job.Type)
		success = true
	}
	return
}

func (c *nomadClient) evaluationInspector(evalID *string) error {

	for {
		evalInfo, _, err := c.nomad.Evaluations().Info(*evalID, nil)
		if err != nil {
			return err
		}

		switch evalInfo.Status {
		case nomadStructs.EvalStatusComplete, nomadStructs.EvalStatusFailed, nomadStructs.EvalStatusCancelled:
			if len(evalInfo.FailedTGAllocs) == 0 {
				logging.Info("levant/deploy: evaluation %s finished successfully", *evalID)
				return nil
			}

			for group, metrics := range evalInfo.FailedTGAllocs {

				// Check if any nodes have been exhausted of resources and therfore are
				// unable to place allocs.
				if metrics.NodesExhausted > 0 {
					var exhausted, dimension []string
					for e := range metrics.ClassExhausted {
						exhausted = append(exhausted, e)
					}
					for d := range metrics.DimensionExhausted {
						dimension = append(dimension, d)
					}
					logging.Error("levant/deploy: task group %s failed to place allocs, failed on %v and exhausted %v",
						group, exhausted, dimension)
				}

				// Check if any node classes were filtered causing alloc placement
				// failures.
				if len(metrics.ClassFiltered) > 0 {
					for f := range metrics.ClassFiltered {
						logging.Error("levant/deploy: task group %s failed to place %v allocs as class \"%s\" was filtered",
							group, len(metrics.ClassFiltered), f)
					}
				}

				// Check if any node constraints were filtered causing alloc placement
				// failures.
				if len(metrics.ConstraintFiltered) > 0 {
					for cf := range metrics.ConstraintFiltered {
						logging.Error("levant/deploy: task group %s failed to place %v allocs as constraint \"%s\" was filtered",
							group, len(metrics.ConstraintFiltered), cf)
					}
				}
			}

			return fmt.Errorf("evaluation %v finished with status %s but failed to place allocations",
				*evalID, evalInfo.Status)

		default:
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (c *nomadClient) deploymentWatcher(depID string, autoPromote int) (success bool) {

	var canaryChan chan interface{}
	deploymentChan := make(chan interface{})

	t := time.Now()
	wt := time.Duration(5 * time.Second)

	// Setup the canaryChan and launch the autoPromote go routine if autoPromote
	// has been enabled.
	if autoPromote > 0 {
		canaryChan = make(chan interface{})
		go c.canaryAutoPromote(depID, autoPromote, canaryChan, deploymentChan)
	}

	q := &nomad.QueryOptions{WaitIndex: 1, AllowStale: true, WaitTime: wt}

	for {

		dep, meta, err := c.nomad.Deployments().Info(depID, q)
		logging.Debug("levant/deploy: deployment %v running for %.2fs", depID, time.Since(t).Seconds())

		// Listen for the deploymentChan closing which indicates Levant should exit
		// the deployment watcher.
		select {
		case <-deploymentChan:
			return false
		default:
			break
		}

		if err != nil {
			logging.Error("levant/deploy: unable to get info of deployment %s: %v", depID, err)
			return
		}

		if meta.LastIndex <= q.WaitIndex {
			continue
		}

		q.WaitIndex = meta.LastIndex

		cont, err := c.checkDeploymentStatus(dep, canaryChan)
		if err != nil {
			return false
		}

		if cont {
			continue
		} else {
			return true
		}
	}
}

func (c *nomadClient) checkDeploymentStatus(dep *nomad.Deployment, shutdownChan chan interface{}) (bool, error) {

	switch dep.Status {
	case nomadStructs.DeploymentStatusSuccessful:
		logging.Info("levant/deploy: deployment %v has completed successfully", dep.ID)
		return false, nil
	case nomadStructs.DeploymentStatusRunning:
		return true, nil
	default:
		if shutdownChan != nil {
			logging.Debug("levant/deploy: deployment %v meaning canary auto promote will shutdown", dep.Status)
			close(shutdownChan)
		}

		logging.Error("levant/deploy: deployment %v has status %s", dep.ID, dep.Status)

		// Launch the failure inspector.
		c.checkFailedDeployment(&dep.ID)

		return false, fmt.Errorf("deployment failed")
	}
}

// canaryAutoPromote handles Levant's canary-auto-promote functionality.
func (c *nomadClient) canaryAutoPromote(depID string, waitTime int, shutdownChan, deploymentChan chan interface{}) {

	// Setup the AutoPromote timer.
	autoPromote := time.After(time.Duration(waitTime) * time.Second)

	for {
		select {
		case <-autoPromote:
			logging.Info("levant/deploy: auto-promote period %vs has been reached for deployment %s",
				waitTime, depID)

			// Check the deployment is healthy before promoting.
			if healthy := c.checkCanaryDeploymentHealth(depID); !healthy {
				logging.Error("levant/deploy: the canary deployment %s has unhealthy allocations, unable to promote", depID)
				close(deploymentChan)
				return
			}

			logging.Info("levant/deploy: triggering auto promote of deployment %s", depID)

			// Promote the deployment.
			_, _, err := c.nomad.Deployments().PromoteAll(depID, nil)
			if err != nil {
				logging.Error("levant/deploy: unable to promote deployment %s: %v", depID, err)
				close(deploymentChan)
				return
			}

		case <-shutdownChan:
			logging.Info("levant/deploy: canary auto promote has been shutdown")
			return
		}
	}
}

// checkCanaryDeploymentHealth is used to check the health status of each
// task-group within a canary deployment.
func (c *nomadClient) checkCanaryDeploymentHealth(depID string) (healthy bool) {

	var unhealthy int

	dep, _, err := c.nomad.Deployments().Info(depID, &nomad.QueryOptions{AllowStale: true})
	if err != nil {
		logging.Error("levant/deploy: unable to query deployment %s for health: %v", depID, err)
		return
	}

	// Itertate each task in the deployment to determine is health status. If an
	// unhealthy task is found, incrament the unhealthy counter.
	for taskName, taskInfo := range dep.TaskGroups {
		// skip any task groups which are not configured for canary deployments
		if taskInfo.DesiredCanaries == 0 {
			logging.Debug("levant/deploy: task %s has no desired canaries, skipping health checks in deployment %s", taskName, depID)
			continue
		}

		if taskInfo.DesiredCanaries != taskInfo.HealthyAllocs {
			logging.Error("levant/deploy: task %s has unhealthy allocations in deployment %s", taskName, depID)
			unhealthy++
		}
	}

	// If zero unhealthy tasks were found, continue with the auto promotion.
	if unhealthy == 0 {
		logging.Debug("levant/deploy: deployment %s has 0 unhealthy allocations", depID)
		healthy = true
	}

	return
}

// getDeploymentID finds the Nomad deploymentID associated to a Nomad
// evaluationID. This is only needed as sometimes Nomad initially returns eval
// info with an empty deploymentID; and a retry is required in order to get the
// updated response from Nomad.
func (c *nomadClient) getDeploymentID(evalID string) (depID string, err error) {

	var evalInfo *nomad.Evaluation

	for {
		if evalInfo, _, err = c.nomad.Evaluations().Info(evalID, nil); err != nil {
			return
		}

		if evalInfo.DeploymentID == "" {
			logging.Debug("levant/deploy: Nomad returned an empty deployment for evaluation %v; retrying", evalID)
			time.Sleep(2 * time.Second)
			continue
		} else {
			break
		}
	}

	return evalInfo.DeploymentID, nil
}

// dynamicGroupCountUpdater takes the templated and rendered job and updates the
// group counts based on the currently deployed job; if its running.
func (c *nomadClient) dynamicGroupCountUpdater(job *nomad.Job) error {

	// Gather information about the current state, if any, of the job on the
	// Nomad cluster.
	rJob, _, err := c.nomad.Jobs().Info(*job.Name, &nomad.QueryOptions{})

	// This is a hack due to GH-1849; we check the error string for 404 which
	// indicates the job is not running, not that there was an error in the API
	// call.
	if err != nil && strings.Contains(err.Error(), "404") {
		logging.Info("levant/deploy: job %s not running, using template file group counts", *job.Name)
		return nil
	} else if err != nil {
		logging.Error("levant/deploy: unable to perform job evaluation: %v", err)
		return err
	}

	// Iterate the templated job and the Nomad returned job and update group count
	// based on matches.
	for _, rGroup := range rJob.TaskGroups {
		for _, group := range job.TaskGroups {
			if *rGroup.Name == *group.Name {
				logging.Info("levant/deploy: using dynamic count %v for job %s and group %s",
					*rGroup.Count, *job.Name, *group.Name)
				group.Count = rGroup.Count
			}
		}
	}
	return nil
}
