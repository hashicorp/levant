package levant

import (
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
	Deploy(*nomad.Job, int) bool
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
func (c *nomadClient) Deploy(job *nomad.Job, autoPromote int) (success bool) {

	// Validate the job to check it is syntactically correct.
	if _, _, err := c.nomad.Jobs().Validate(job, nil); err != nil {
		logging.Error("levant/deploy: job validation failed: %v", err)
		return
	}

	logging.Debug("levant/deploy: running dynamic job count updater for job %s", *job.Name)
	if err := c.dynamicGroupCountUpdater(job); err != nil {
		return
	}

	logging.Info("levant/deploy: triggering a deployment of job %s", *job.Name)

	eval, _, err := c.nomad.Jobs().Register(job, nil)
	if err != nil {
		logging.Error("levant/deploy: unable to register job %s with Nomad: %v", *job.Name, err)
		return
	}

	switch *job.Type {
	case nomadStructs.JobTypeService:
		logging.Debug("levant/deploy: beginning deployment watcher for job %s", *job.Name)
		success = c.deploymentWatcher(eval.EvalID, autoPromote)
	case nomadStructs.JobTypeBatch, nomadStructs.JobTypeSystem:
		logging.Debug("levant/deploy: job type %s does not support Nomad deployment model", *job.Type)
		success = true
	}

	return
}

func (c *nomadClient) deploymentWatcher(evalID string, autoPromote int) (success bool) {

	var shutdownChan chan interface{}

	t := time.Now()
	wt := time.Duration(5 * time.Second)

	// Get the deploymentID from the evaluationID so that we can watch the
	// deployment for end status.
	depID, err := c.getDeploymentID(evalID)
	if err != nil {
		logging.Error("levant/deploy: unable to get info of evaluation %s: %v", evalID, err)
	}

	if autoPromote > 0 {
		shutdownChan = make(chan interface{})
		go c.canaryAutoPromote(depID, autoPromote, shutdownChan)
	}

	q := &nomad.QueryOptions{WaitIndex: 1, AllowStale: true, WaitTime: wt}

	for {
		dep, meta, err := c.nomad.Deployments().Info(depID, q)
		logging.Info("levant/deploy: deployment %v running for %v", depID, time.Since(t))

		if err != nil {
			logging.Error("levant/deploy: unable to get info of deployment %s: %v", depID, err)
			return
		}

		if meta.LastIndex <= q.WaitIndex {
			continue
		}

		q.WaitIndex = meta.LastIndex

		switch dep.Status {
		case nomadStructs.DeploymentStatusSuccessful:
			success = true
			logging.Info("levant/deploy: deployment %v succeeded in %v", depID, time.Since(t))
			return
		case nomadStructs.DeploymentStatusRunning:
			continue
		default:
			if shutdownChan != nil {
				logging.Info("levant/deploy: canary auto promote has been asked to shutdown")
				close(shutdownChan)
			}

			success = false
			c.checkFailedDeployment(&depID)
			logging.Error("levant/deploy: deployment %v failed in %v", depID, time.Since(t))
			return
		}
	}
}

// canaryAutoPromote handles Levant's canary-auto-promote functionality.
func (c *nomadClient) canaryAutoPromote(depID string, waitTime int, shutdownChan chan interface{}) {

	// Setup the AutoPromote timer.
	autoPromote := time.After(time.Duration(waitTime) * time.Second)

	for {
		select {
		case <-autoPromote:
			logging.Info("levant/deploy: auto-promote period %v has been reached for deployment %s",
				autoPromote, depID)

			// Check the deployment is healthy before promoting.
			if healthly := c.checkCanaryDeploymentHealth(depID); !healthly {
				return
			}

			logging.Info("levant/deploy: triggering auto promote of deployment %s", depID)

			// Promote the deployment.
			_, _, err := c.nomad.Deployments().PromoteAll(depID, nil)
			if err != nil {
				logging.Error("levant/deploy: unable to promote deployment %s: %v", depID, err)
			}

		case <-shutdownChan:
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
		if taskInfo.DesiredCanaries != taskInfo.PlacedAllocs && taskInfo.DesiredCanaries != taskInfo.HealthyAllocs {
			logging.Error("levant/deploy: task %s has unhealthy allocations in deployment %s", taskName, depID)
			unhealthy++
		}
	}

	// If zero unhealthy tasks were found, continue with the auto promotion.
	if unhealthy == 0 {
		logging.Info("levant/deploy: deployment %s has 0 unhealthy allocations", depID)
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
