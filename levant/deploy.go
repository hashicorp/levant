// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package levant

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/levant/client"
	"github.com/hashicorp/levant/levant/structs"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	jobStatusRunning = "running"
)

// levantDeployment is the all deployment related objects for this Levant
// deployment invocation.
type levantDeployment struct {
	nomad   *nomad.Client
	config  *DeployConfig
	options *nomad.WriteOptions
}

// DeployConfig is the set of config structs required to run a Levant deploy.
type DeployConfig struct {
	Deploy   *structs.DeployConfig
	Client   *structs.ClientConfig
	Plan     *structs.PlanConfig
	Template *structs.TemplateConfig
}

// newLevantDeployment sets up the Levant deployment object and Nomad client
// to interact with the Nomad API.
func newLevantDeployment(config *DeployConfig, nomadClient *nomad.Client) (*levantDeployment, error) {

	var err error
	if config.Deploy.EnvVault {
		config.Deploy.VaultToken = os.Getenv("VAULT_TOKEN")
	}

	dep := &levantDeployment{}
	dep.config = config
	dep.options = setWriteOptions(dep.config.Template)

	if nomadClient == nil {
		dep.nomad, err = client.NewNomadClient(config.Client.Addr)
		if err != nil {
			return nil, err
		}
	} else {
		dep.nomad = nomadClient
	}

	// Add the JobID as a log context field.
	log.Logger = log.With().Str(structs.JobIDContextField, *config.Template.Job.ID).Logger()

	return dep, nil
}

// TriggerDeployment provides the main entry point into a Levant deployment and
// is used to setup the clients before triggering the deployment process.
func TriggerDeployment(config *DeployConfig, nomadClient *nomad.Client) bool {

	// Create our new deployment object.
	levantDep, err := newLevantDeployment(config, nomadClient)
	if err != nil {
		log.Error().Err(err).Msg("levant/deploy: unable to setup Levant deployment")
		return false
	}

	// Run the job validation steps and count updater.
	preDepVal := levantDep.preDeployValidate()
	if !preDepVal {
		log.Error().Msg("levant/deploy: pre-deployment validation process failed")
		return false
	}

	// Start the main deployment function.
	success := levantDep.deploy()
	if !success {
		log.Error().Msg("levant/deploy: job deployment failed")
		return false
	}

	log.Info().Msg("levant/deploy: job deployment successful")
	return true
}

func (l *levantDeployment) preDeployValidate() (success bool) {

	// Validate the job to check it is syntactically correct.
	if _, _, err := l.nomad.Jobs().Validate(l.config.Template.Job, l.options); err != nil {
		log.Error().Err(err).Msg("levant/deploy: job validation failed")
		return
	}

	// If job.Type isn't set we can't continue
	if l.config.Template.Job.Type == nil {
		log.Error().Msgf("levant/deploy: Nomad job `type` is not set; should be set to `%s`, `%s` or `%s`",
			nomad.JobTypeBatch, nomad.JobTypeSystem, nomad.JobTypeService)
		return
	}

	if !l.config.Deploy.ForceCount {
		if err := l.dynamicGroupCountUpdater(); err != nil {
			return
		}
	}

	return true
}

// deploy triggers a register of the job resulting in a Nomad deployment which
// is monitored to determine the eventual state.
func (l *levantDeployment) deploy() (success bool) {

	log.Info().Msgf("levant/deploy: triggering a deployment")

	l.config.Template.Job.VaultToken = &l.config.Deploy.VaultToken

	eval, _, err := l.nomad.Jobs().Register(l.config.Template.Job, l.options)
	if err != nil {
		log.Error().Err(err).Msg("levant/deploy: unable to register job with Nomad")
		return
	}

	if l.config.Deploy.ForceBatch {
		if eval.EvalID, err = l.triggerPeriodic(l.config.Template.Job.ID); err != nil {
			log.Error().Err(err).Msg("levant/deploy: unable to trigger periodic instance of job")
			return
		}
	}

	// Periodic and parameterized jobs do not return an evaluation and therefore
	// can't perform the evaluationInspector unless we are forcing an instance of
	// periodic which will yield an EvalID.
	if !l.config.Template.Job.IsPeriodic() && !l.config.Template.Job.IsParameterized() ||
		l.config.Template.Job.IsPeriodic() && l.config.Deploy.ForceBatch {

		// Trigger the evaluationInspector to identify any potential errors in the
		// Nomad evaluation run. As far as I can tell from testing; a single alloc
		// failure in an evaluation means no allocs will be placed so we exit here.
		err = l.evaluationInspector(&eval.EvalID)
		if err != nil {
			log.Error().Err(err).Msg("levant/deploy: something")
			return
		}
	}

	if l.isJobZeroCount() {
		return true
	}

	switch *l.config.Template.Job.Type {
	case nomad.JobTypeService:

		// If the service job doesn't have an update stanza, the job will not use
		// Nomad deployments.
		if l.config.Template.Job.Update == nil {
			log.Info().Msg("levant/deploy: job is not configured with update stanza, consider adding to use deployments")
			return l.jobStatusChecker(&eval.EvalID)
		}

		log.Info().Msgf("levant/deploy: beginning deployment watcher for job")

		// Get the deploymentID from the evaluationID so that we can watch the
		// deployment for end status.
		depID, err := l.getDeploymentID(eval.EvalID)
		if err != nil {
			log.Error().Err(err).Msgf("levant/deploy: unable to get info of evaluation %s", eval.EvalID)
			return
		}

		// Get the success of the deployment and return if we have success.
		if success = l.deploymentWatcher(depID); success {
			return
		}

		dep, _, err := l.nomad.Deployments().Info(depID, setQueryOptions(l.options))
		if err != nil {
			log.Error().Err(err).Msgf("levant/deploy: unable to query deployment %s for auto-revert check", depID)
			return
		}

		// If the job is not a canary job, then run the auto-revert checker, the
		// current checking mechanism is slightly hacky and should be updated.
		// The reason for this is currently the config.Job is populate from the
		// rendered job and so a user could potentially not set canary meaning
		// the field shows a null.
		if l.config.Template.Job.Update.Canary == nil {
			l.checkAutoRevert(dep)
		} else if *l.config.Template.Job.Update.Canary == 0 {
			l.checkAutoRevert(dep)
		}

	case nomad.JobTypeBatch:
		return l.jobStatusChecker(&eval.EvalID)

	case nomad.JobTypeSystem:
		return l.jobStatusChecker(&eval.EvalID)

	default:
		log.Debug().Msgf("levant/deploy: Levant does not support advanced deployments of job type %s",
			*l.config.Template.Job.Type)
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
		case "complete", "failed", "canceled":
			if len(evalInfo.FailedTGAllocs) == 0 {
				log.Info().Msgf("levant/deploy: evaluation %s finished successfully", *evalID)
				return nil
			}

			for group, metrics := range evalInfo.FailedTGAllocs {

				// Check if any nodes have been exhausted of resources and therefor are
				// unable to place allocs.
				if metrics.NodesExhausted > 0 {
					var exhausted, dimension []string
					for e := range metrics.ClassExhausted {
						exhausted = append(exhausted, e)
					}
					for d := range metrics.DimensionExhausted {
						dimension = append(dimension, d)
					}
					log.Error().Msgf("levant/deploy: task group %s failed to place allocs, failed on %v and exhausted %v",
						group, exhausted, dimension)
				}

				// Check if any node classes were filtered causing alloc placement
				// failures.
				if len(metrics.ClassFiltered) > 0 {
					for f := range metrics.ClassFiltered {
						log.Error().Msgf("levant/deploy: task group %s failed to place %v allocs as class \"%s\" was filtered",
							group, len(metrics.ClassFiltered), f)
					}
				}

				// Check if any node constraints were filtered causing alloc placement
				// failures.
				if len(metrics.ConstraintFiltered) > 0 {
					for cf := range metrics.ConstraintFiltered {
						log.Error().Msgf("levant/deploy: task group %s failed to place %v allocs as constraint \"%s\" was filtered",
							group, len(metrics.ConstraintFiltered), cf)
					}
				}
			}

			// Do not return an error here; there could well be information from
			// Nomad detailing filtered nodes but the deployment will still be
			// successful. GH-220.
			return nil

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
	wt := 5 * time.Second

	// Setup the canaryChan and launch the autoPromote go routine if autoPromote
	// has been enabled.
	if l.config.Deploy.Canary > 0 {
		canaryChan = make(chan interface{})
		go l.canaryAutoPromote(depID, l.config.Deploy.Canary, canaryChan, deploymentChan)
	}

	q := &nomad.QueryOptions{WaitIndex: 1, AllowStale: l.config.Client.AllowStale, WaitTime: wt}

	for {

		dep, meta, err := l.nomad.Deployments().Info(depID, q)
		log.Debug().Msgf("levant/deploy: deployment %v running for %.2fs", depID, time.Since(t).Seconds())

		// Listen for the deploymentChan closing which indicates Levant should exit
		// the deployment watcher.
		select {
		case <-deploymentChan:
			return false
		default:
			break
		}

		if err != nil {
			log.Error().Err(err).Msgf("levant/deploy: unable to get info of deployment %s", depID)
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
	case "successful":
		log.Info().Msgf("levant/deploy: deployment %v has completed successfully", dep.ID)
		return false, nil
	case jobStatusRunning:
		return true, nil
	default:
		if shutdownChan != nil {
			log.Debug().Msgf("levant/deploy: deployment %v meaning canary auto promote will shutdown", dep.Status)
			close(shutdownChan)
		}

		log.Error().Msgf("levant/deploy: deployment %v has status %s", dep.ID, dep.Status)

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
			log.Info().Msgf("levant/deploy: auto-promote period %vs has been reached for deployment %s",
				waitTime, depID)

			// Check the deployment is healthy before promoting.
			if healthy := l.checkCanaryDeploymentHealth(depID); !healthy {
				log.Error().Msgf("levant/deploy: the canary deployment %s has unhealthy allocations, unable to promote", depID)
				close(deploymentChan)
				return
			}

			log.Info().Msgf("levant/deploy: triggering auto promote of deployment %s", depID)

			// Promote the deployment.
			_, _, err := l.nomad.Deployments().PromoteAll(depID, nil)
			if err != nil {
				log.Error().Err(err).Msgf("levant/deploy: unable to promote deployment %s", depID)
				close(deploymentChan)
				return
			}

		case <-shutdownChan:
			log.Info().Msg("levant/deploy: canary auto promote has been shutdown")
			return
		}
	}
}

// checkCanaryDeploymentHealth is used to check the health status of each
// task-group within a canary deployment.
func (l *levantDeployment) checkCanaryDeploymentHealth(depID string) (healthy bool) {

	var unhealthy int

	dep, _, err := l.nomad.Deployments().Info(depID, &nomad.QueryOptions{AllowStale: l.config.Client.AllowStale})
	if err != nil {
		log.Error().Err(err).Msgf("levant/deploy: unable to query deployment %s for health", depID)
		return
	}

	// Iterate over each task in the deployment to determine its health status. If an
	// unhealthy task is found, increment the unhealthy counter.
	for taskName, taskInfo := range dep.TaskGroups {
		// skip any task groups which are not configured for canary deployments
		if taskInfo.DesiredCanaries == 0 {
			log.Debug().Msgf("levant/deploy: task %s has no desired canaries, skipping health checks in deployment %s", taskName, depID)
			continue
		}

		if taskInfo.DesiredCanaries != taskInfo.HealthyAllocs {
			log.Error().Msgf("levant/deploy: task %s has unhealthy allocations in deployment %s", taskName, depID)
			unhealthy++
		}
	}

	// If zero unhealthy tasks were found, continue with the auto promotion.
	if unhealthy == 0 {
		log.Debug().Msgf("levant/deploy: deployment %s has 0 unhealthy allocations", depID)
		healthy = true
	}

	return
}

// triggerPeriodic is used to force an instance of a periodic job outside of the
// planned schedule. This results in an evalID being created that can then be
// checked in the same fashion as other jobs.
func (l *levantDeployment) triggerPeriodic(jobID *string) (evalID string, err error) {

	log.Info().Msg("levant/deploy: triggering a run of periodic job")

	// Trigger the run if possible and just return both the evalID and the err.
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

	timeout := time.NewTicker(time.Second * 60)
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			err = errors.New("timeout reached on attempting to find deployment ID")
			return

		default:
			if evalInfo, _, err = l.nomad.Evaluations().Info(evalID, nil); err != nil {
				return
			}

			if evalInfo.DeploymentID != "" {
				return evalInfo.DeploymentID, nil
			}

			log.Debug().Msgf("levant/deploy: Nomad returned an empty deployment for evaluation %v; retrying", evalID)
			time.Sleep(2 * time.Second)
			continue
		}
	}
}

// dynamicGroupCountUpdater takes the templated and rendered job and updates the
// group counts based on the currently deployed job; if it's running.
func (l *levantDeployment) dynamicGroupCountUpdater() error {

	// Gather information about the current state, if any, of the job on the
	// Nomad cluster.
	rJob, _, err := l.nomad.Jobs().Info(*l.config.Template.Job.Name, &nomad.QueryOptions{})

	// This is a hack due to GH-1849; we check the error string for 404, which
	// indicates the job is not running, not that there was an error in the API
	// call.
	if err != nil && strings.Contains(err.Error(), "404") {
		log.Info().Msg("levant/deploy: job is not running, using template file group counts")
		return nil
	} else if err != nil {
		log.Error().Err(err).Msg("levant/deploy: unable to perform job evaluation")
		return err
	}

	// Check that the job is actually running and not in a potentially stopped
	// state.
	if *rJob.Status != jobStatusRunning {
		return nil
	}

	log.Debug().Msgf("levant/deploy: running dynamic job count updater")

	// Iterate over the templated job and the Nomad returned job and update group count
	// based on matches.
	for _, rGroup := range rJob.TaskGroups {
		for _, group := range l.config.Template.Job.TaskGroups {
			if *rGroup.Name == *group.Name {
				log.Info().Msgf("levant/deploy: using dynamic count %v for group %s",
					*rGroup.Count, *group.Name)
				group.Count = rGroup.Count
			}
		}
	}
	return nil
}

// isJobZeroCount checks that all task groups have a count bigger than zero.
func (l *levantDeployment) isJobZeroCount() bool {
	for _, tg := range l.config.Template.Job.TaskGroups {
		if tg.Count == nil {
			return false
		} else if *tg.Count > 0 {
			return false
		}
	}
	return true
}
