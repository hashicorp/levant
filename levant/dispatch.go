package levant

import (
	"github.com/hashicorp/levant/client"
	"github.com/hashicorp/levant/levant/structs"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog/log"
)

// TriggerDispatch provides the main entry point into a Levant dispatch and
// is used to setup the clients before triggering the dispatch process.
func TriggerDispatch(job string, metaMap map[string]string, payload []byte, address string) bool {

	client, err := client.NewNomadClient(address)
	if err != nil {
		log.Error().Msgf("levant/dispatch: unable to setup Levant dispatch: %v", err)
		return false
	}

	// TODO: Potential refactor so that dispatch does not need to use the
	// levantDeployment object. Requires client refactor.
	dep := &levantDeployment{}
	dep.nomad = client

	success := dep.dispatch(job, metaMap, payload)
	if !success {
		log.Error().Msgf("levant/dispatch: dispatch of job %v failed", job)
		return false
	}

	log.Info().Msgf("levant/dispatch: dispatch of job %v successful", job)
	return true
}

// dispatch triggers a new instance of a parameterized job of the job
// resulting in a Nomad job which is monitored to determine the eventual
// state.
func (l *levantDeployment) dispatch(job string, metaMap map[string]string, payload []byte) bool {

	// Initiate the dispatch with the passed meta parameters.
	eval, _, err := l.nomad.Jobs().Dispatch(job, metaMap, payload, nil)
	if err != nil {
		log.Error().Msgf("levant/dispatch: %v", err)
		return false
	}

	log.Info().Msgf("levant/dispatch: triggering dispatch against job %s", job)

	// If we didn't get an EvaluationID then we cannot continue.
	if eval.EvalID == "" {
		log.Error().Msgf("levant/dispatch: dispatched job %s did not return evaluation", job)
		return false
	}

	// In order to correctly run the jobStatusChecker we need to correctly
	// assign the dispatched job ID/Name based on the invoked job.
	l.config = &DeployConfig{
		Template: &structs.TemplateConfig{
			Job: &nomad.Job{
				ID:   &eval.DispatchedJobID,
				Name: &eval.DispatchedJobID,
			},
		},
	}

	// Perform the evaluation inspection to ensure to check for any possible
	// errors in triggering the dispatch job.
	err = l.evaluationInspector(&eval.EvalID)
	if err != nil {
		log.Error().Msgf("levant/dispatch: %v", err)
		return false
	}

	return l.jobStatusChecker(&eval.EvalID)
}
