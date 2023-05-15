// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scale

import (
	"github.com/hashicorp/levant/client"
	"github.com/hashicorp/levant/levant"
	"github.com/hashicorp/levant/levant/structs"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog/log"
)

// Config is the set of config structs required to run a Levant scale.
type Config struct {
	Client *structs.ClientConfig
	Scale  *structs.ScaleConfig
}

// TriggerScalingEvent provides the exported entry point into performing a job
// scale based on user inputs.
func TriggerScalingEvent(config *Config) bool {

	// Add the JobID as a log context field.
	log.Logger = log.With().Str(structs.JobIDContextField, config.Scale.JobID).Logger()

	nomadClient, err := client.NewNomadClient(config.Client.Addr)
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
	deploymentConfig := &levant.DeployConfig{}
	deploymentConfig.Template = &structs.TemplateConfig{Job: job}
	deploymentConfig.Client = config.Client
	deploymentConfig.Deploy = &structs.DeployConfig{ForceCount: true}

	log.Info().Msg("levant/scale: job will now be deployed with updated counts")

	// Trigger a deployment of the updated job which results in the scaling of
	// the job and will go through all the deployment tracking until an end
	// state is reached.
	return levant.TriggerDeployment(deploymentConfig, nomadClient)
}

// updateJob gathers information on the current state of the running job and
// along with the user defined input updates the in-memory job specification
// to reflect the desired scaled state.
func updateJob(client *nomad.Client, config *Config) *nomad.Job {

	job, _, err := client.Jobs().Info(config.Scale.JobID, nil)
	if err != nil {
		log.Error().Err(err).Msg("levant/scale: unable to obtain job information from Nomad")
		return nil
	}

	// You can't scale a job that isn't running; or at least you shouldn't in
	// my current opinion.
	if *job.Status != "running" {
		log.Error().Msgf("levant/scale: job is not in running state")
		return nil
	}

	for _, group := range job.TaskGroups {

		// If the user has specified a taskgroup to scale, ensure we only change
		// the specific of this.
		if config.Scale.TaskGroup != "" {
			if *group.Name == config.Scale.TaskGroup {
				log.Debug().Msgf("levant/scale: scaling action to be requested on taskgroup %s only",
					config.Scale.TaskGroup)
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
func updateTaskGroup(config *Config, group *nomad.TaskGroup) {

	var c int

	// If a percentage scale value has been passed, we must convert this to an
	// int which represents the count to scale by as Nomad job submissions must
	// be done with group counts as desired ints.
	switch config.Scale.DirectionType {
	case structs.ScalingDirectionTypeCount:
		c = config.Scale.Count
	case structs.ScalingDirectionTypePercent:
		c = calculateCountBasedOnPercent(*group.Count, config.Scale.Percent)
	}

	// Depending on whether we are scaling-out or scaling-in we need to perform
	// the correct maths. There is a little duplication here, but that is to
	// provide better logging.
	switch config.Scale.Direction {
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
