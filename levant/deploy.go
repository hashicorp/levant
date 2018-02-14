package levant

import (
	"fmt"
	"strings"
	"time"

	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/jrasell/levant/levant/structs"
	"github.com/jrasell/levant/logging"
)

// levantDeployment is the all deployment related objects for this Levant
// deployment invoction.
type levantDeployment struct {
	nomad  *nomad.Client
	config *structs.Config
}

// newNomadClient is used to create a new client to interact with Nomad.
func newNomadClient(addr string) (*nomad.Client, error) {
	config := nomad.DefaultConfig()

	if addr != "" {
		config.Address = addr
	}

	c, err := nomad.NewClient(config)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// newLevantDeployment sets up the Levant deployment object and Nomad client
// to interact with the Nomad API.
func newLevantDeployment(config *structs.Config) (*levantDeployment, error) {

	var err error

	dep := &levantDeployment{}
	dep.config = config

	dep.nomad, err = newNomadClient(config.Addr)
	if err != nil {
		return nil, err
	}

	return dep, nil
}

// TriggerDeployment provides the main entry point into a Levant deployment and
// is used to setup the clients before triggering the deployment process.
func TriggerDeployment(config *structs.Config) bool {

	// Create our new deployment object.
	levantDep, err := newLevantDeployment(config)
	if err != nil {
		logging.Error("levant/deploy: unable to setup Levant deployment: %v", err)
		return false
	}

	// Start the main deployment function.
	success := levantDep.deploy()
	if !success {
		logging.Error("levant/deploy: deployment of job %v failed", *config.Job.Name)
		return false
	}

	logging.Info("levant/deploy: deployment of job %v successful", *config.Job.Name)
	return true
}

// deploy triggers a register of the job resulting in a Nomad deployment which
// is monitored to determine the eventual state.
func (l *levantDeployment) deploy() (success bool) {

	// Validate the job to check it is syntactically correct.
	if _, _, err := l.nomad.Jobs().Validate(l.config.Job, nil); err != nil {
		logging.Error("levant/deploy: job validation failed: %v", err)
		return
	}

	// If job.Type isn't set we can't continue
	if l.config.Job.Type == nil {
		logging.Error("levant/deploy: Nomad job `type` is not set; should be set to `%s`, `%s` or `%s`",
			nomadStructs.JobTypeBatch, nomadStructs.JobTypeSystem, nomadStructs.JobTypeService)
		return
	}

	if !l.config.ForceCount {
		if err := l.dynamicGroupCountUpdater(); err != nil {
			return
		}
	}

	logging.Info("levant/deploy: triggering a deployment of job %s", *l.config.Job.Name)

	eval, _, err := l.nomad.Jobs().Register(l.config.Job, nil)
	if err != nil {
		logging.Error("levant/deploy: unable to register job %s with Nomad: %v", *l.config.Job.Name, err)
		return
	}

	if l.config.ForceBatch {
		if eval.EvalID, err = l.triggerPeriodic(l.config.Job.ID); err != nil {
			logging.Error("levant/deploy: unable to trigger periodic instance of job %s: %v",
				*l.config.Job.Name, err)
			return
		}
	}

	// Periodic and parameterized jobs do not return an evaluation and therefore
	// can't perform the evaluationInspector unless we are forcing an instance of
	// periodic which will yeild an EvalID.
	if !l.config.Job.IsPeriodic() && !l.config.Job.IsParameterized() ||
		l.config.Job.IsPeriodic() && l.config.ForceBatch {

		// Trigger the evaluationInspector to identify any potential errors in the
		// Nomad evaluation run. As far as I can tell from testing; a single alloc
		// failure in an evaluation means no allocs will be placed so we exit here.
		err = l.evaluationInspector(&eval.EvalID)
		if err != nil {
			logging.Error("levant/deploy: %v", err)
			return
		}
	}

	switch *l.config.Job.Type {
	case nomadStructs.JobTypeService:

		// If the service job doesn't have an update stanza, the job will not use
		// Nomad deployments.
		if l.config.Job.Update == nil {
			logging.Info("levant/deploy: job %s is not configured with update stanza, consider adding to use deployments",
				*l.config.Job.Name)
			return l.jobStatusChecker(&eval.EvalID)
		}

		logging.Info("levant/deploy: beginning deployment watcher for job %s", *l.config.Job.Name)

		// Get the deploymentID from the evaluationID so that we can watch the
		// deployment for end status.
		depID, err := l.getDeploymentID(eval.EvalID)
		if err != nil {
			logging.Error("levant/deploy: unable to get info of evaluation %s: %v", eval.EvalID, err)
			return
		}

		// Get the success of the deployment.
		success = l.deploymentWatcher(depID)

		// If the deployment has not been successful; check whether the job is
		// configured to auto-revert so that this can be tracked.
		if !success {
			dep, _, err := l.nomad.Deployments().Info(depID, nil)
			if err != nil {
				logging.Error("levant/deploy: unable to query deployment %s for auto-revert check: %v",
					dep.ID, err)
				break
			}
			l.checkAutoRevert(dep)
		}

	case nomadStructs.JobTypeBatch:
		return l.jobStatusChecker(&eval.EvalID)

	case nomadStructs.JobTypeSystem:
		return l.jobStatusChecker(&eval.EvalID)

	default:
		logging.Debug("levant/deploy: Levant does not support advanced deployments of job type %s",
			*l.config.Job.Type)
		success = true
	}
	return
}

func (l *levantDeployment) evaluationInspector(evalID *string) error {

	for {
		evalInfo, _, err := l.nomad.Evaluations().Info(*evalID, nil)
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

func (l *levantDeployment) deploymentWatcher(depID string) (success bool) {

	var canaryChan chan interface{}
	deploymentChan := make(chan interface{})

	t := time.Now()
	wt := time.Duration(5 * time.Second)

	// Setup the canaryChan and launch the autoPromote go routine if autoPromote
	// has been enabled.
	if l.config.Canary > 0 {
		canaryChan = make(chan interface{})
		go l.canaryAutoPromote(depID, l.config.Canary, canaryChan, deploymentChan)
	}

	q := &nomad.QueryOptions{WaitIndex: 1, AllowStale: true, WaitTime: wt}

	for {

		dep, meta, err := l.nomad.Deployments().Info(depID, q)
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

		cont, err := l.checkDeploymentStatus(dep, canaryChan)
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

func (l *levantDeployment) checkDeploymentStatus(dep *nomad.Deployment, shutdownChan chan interface{}) (bool, error) {

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
		l.checkFailedDeployment(&dep.ID)

		return false, fmt.Errorf("deployment failed")
	}
}

// canaryAutoPromote handles Levant's canary-auto-promote functionality.
func (l *levantDeployment) canaryAutoPromote(depID string, waitTime int, shutdownChan, deploymentChan chan interface{}) {

	// Setup the AutoPromote timer.
	autoPromote := time.After(time.Duration(waitTime) * time.Second)

	for {
		select {
		case <-autoPromote:
			logging.Info("levant/deploy: auto-promote period %vs has been reached for deployment %s",
				waitTime, depID)

			// Check the deployment is healthy before promoting.
			if healthy := l.checkCanaryDeploymentHealth(depID); !healthy {
				logging.Error("levant/deploy: the canary deployment %s has unhealthy allocations, unable to promote", depID)
				close(deploymentChan)
				return
			}

			logging.Info("levant/deploy: triggering auto promote of deployment %s", depID)

			// Promote the deployment.
			_, _, err := l.nomad.Deployments().PromoteAll(depID, nil)
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
func (l *levantDeployment) checkCanaryDeploymentHealth(depID string) (healthy bool) {

	var unhealthy int

	dep, _, err := l.nomad.Deployments().Info(depID, &nomad.QueryOptions{AllowStale: true})
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

// triggerPeriodic is used to force an instance of a periodic job outside of the
// planned schedule. This results in an evalID being created that can then be
// checked in the same fashion as other jobs.
func (l *levantDeployment) triggerPeriodic(jobID *string) (evalID string, err error) {

	logging.Info("levant/deploy: triggering a run of periodic job %s", *jobID)

	// Trigger the run if possible and just returning both the evalID and the err.
	// There is no need to check this here as the caller does this.
	evalID, _, err = l.nomad.Jobs().PeriodicForce(*jobID, nil)
	return
}

// getDeploymentID finds the Nomad deploymentID associated to a Nomad
// evaluationID. This is only needed as sometimes Nomad initially returns eval
// info with an empty deploymentID; and a retry is required in order to get the
// updated response from Nomad.
func (l *levantDeployment) getDeploymentID(evalID string) (depID string, err error) {

	var evalInfo *nomad.Evaluation

	for {
		if evalInfo, _, err = l.nomad.Evaluations().Info(evalID, nil); err != nil {
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
func (l *levantDeployment) dynamicGroupCountUpdater() error {

	// Gather information about the current state, if any, of the job on the
	// Nomad cluster.
	rJob, _, err := l.nomad.Jobs().Info(*l.config.Job.Name, &nomad.QueryOptions{})

	// This is a hack due to GH-1849; we check the error string for 404 which
	// indicates the job is not running, not that there was an error in the API
	// call.
	if err != nil && strings.Contains(err.Error(), "404") {
		logging.Info("levant/deploy: job %s not running, using template file group counts", *l.config.Job.Name)
		return nil
	} else if err != nil {
		logging.Error("levant/deploy: unable to perform job evaluation: %v", err)
		return err
	}

	// Check that the job is actually running and not in a potentially stopped
	// state.
	if *rJob.Status != nomadStructs.JobStatusRunning {
		return nil
	}

	logging.Debug("levant/deploy: running dynamic job count updater for job %s", *l.config.Job.Name)

	// Iterate the templated job and the Nomad returned job and update group count
	// based on matches.
	for _, rGroup := range rJob.TaskGroups {
		for _, group := range l.config.Job.TaskGroups {
			if *rGroup.Name == *group.Name {
				logging.Info("levant/deploy: using dynamic count %v for job %s and group %s",
					*rGroup.Count, *l.config.Job.Name, *group.Name)
				group.Count = rGroup.Count
			}
		}
	}
	return nil
}
