package scale

import (
	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"

	"github.com/jrasell/levant/client"
	"github.com/jrasell/levant/levant"
	"github.com/jrasell/levant/levant/structs"
	"github.com/rs/zerolog/log"
)

// TriggerScalingEvent provides the exported entry point into performing a job
// scale based on user inputs.
func TriggerScalingEvent(config *structs.ScalingConfig) bool {

	// Add the JobID as a log context field.
	log.Logger = log.With().Str(structs.JobIDContextField, config.JobID).Logger()

	nomadClient, err := client.NewNomadClient(config.Addr)
	if err != nil {
		log.Error().Msg("levant/scale: unable to setup Levant scaling event")
		return false
	}

	job := updateJob(nomadClient, config)
	if job == nil {
		log.Error().Msg("levant/scale: unable to perform job count update")
		return false
	}

	// Setup a deployment object, as a scaling event is a deployment and should
	// go through the same process and code upgrades.
	deploymentConfig := &structs.Config{}
	deploymentConfig.Job = job
	deploymentConfig.ForceCount = true

	log.Info().Msg("levant/scale: job will now be deployed with updated counts")

	// Trigger a deployment of the updated job which results in the scaling of
	// the job and will go through all the deployment tracking until an end
	// state is reached.
	success := levant.TriggerDeployment(deploymentConfig, nomadClient)
	if !success {
		return false
	}

	return true
}

// updateJob gathers information on the current state of the running job and
// along with the user defined input updates the in-memory job specification
// to reflect the desired scaled state.
func updateJob(client *nomad.Client, config *structs.ScalingConfig) *nomad.Job {

	job, _, err := client.Jobs().Info(config.JobID, nil)
	if err != nil {
		log.Error().Err(err).Msg("levant/scale: unable to obtain job information from Nomad")
		return nil
	}

	// You can't scale a job that isn't running; or at least you shouldn't in
	// my current opinion.
	if *job.Status != nomadStructs.JobStatusRunning {
		log.Error().Msgf("levant/scale: job is not in %s state", nomadStructs.JobStatusRunning)
		return nil
	}

	for _, group := range job.TaskGroups {

		// If the user has specified a taskgroup to scale, ensure we only change
		// the specific of this.
		if config.TaskGroup != "" {
			if *group.Name == config.TaskGroup {
				log.Debug().Msgf("levant/scale: scaling action to be requested on taskgroup %s only",
					config.TaskGroup)
				updateTaskGroup(config, group)
			}

			// If no taskgroup has been specified, all found will have their
			// count updated.
		} else {
			log.Debug().Msg("levant/scale: scaling action requested on all taskgroups")
			updateTaskGroup(config, group)
		}
	}

	return job
}

// updateTaskGroup is tasked with performing the count update based on the user
// configuration when a group is identified as being marked for scaling.
func updateTaskGroup(config *structs.ScalingConfig, group *nomad.TaskGroup) {

	var c int

	// If a percentage scale value has been passed, we must convert this to an
	// int which represents the count to scale by as Nomad job submissions must
	// be done with group counts as desired ints.
	switch config.DirectionType {
	case structs.ScalingDirectionTypeCount:
		c = config.Count
	case structs.ScalingDirectionTypePercent:
		c = calculateCountBasedOnPercent(*group.Count, config.Percent)
	}

	// Depending on whether we are scaling-out or scaling-in we need to perform
	// the correct maths. There is a little duplication here, but that is to
	// provide better logging.
	switch config.Direction {
	case structs.ScalingDirectionOut:
		nc := *group.Count + c
		log.Info().Msgf("levant/scale: task group %s will scale-out from %v to %v",
			*group.Name, *group.Count, nc)
		*group.Count = nc

	case structs.ScalingDirectionIn:
		nc := *group.Count - c
		log.Info().Msgf("levant/scale: task group %s will scale-in from %v to %v",
			*group.Name, *group.Count, nc)
		*group.Count = nc
	}
}

// calculateCountBasedOnPercent is a small helper function to turn a percentage
// based scale event into a relative count.
func calculateCountBasedOnPercent(count, percent int) int {
	n := (float64(count) / 100) * float64(percent)
	return int(n + 0.5)
}
