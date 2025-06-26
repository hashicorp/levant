// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package levant

import (
	"time"

	"github.com/hashicorp/nomad/api"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog/log"
)

func (l *levantDeployment) autoRevert(dep *nomad.Deployment) {

	// Setup a loop in order to retry a race condition whereby Levant may query
	// the latest deployment (auto-revert dep) before it has been started.
	i := 0
	for i := 0; i < 5; i++ {
		revertDep, _, err := l.nomad.Jobs().LatestDeployment(dep.JobID, &api.QueryOptions{Namespace: dep.Namespace})
		if err != nil {
			log.Error().Msgf("levant/auto_revert: unable to query latest deployment of job %s", dep.JobID)
			return
		}

		// Check whether we have got the original deployment ID as a return from
		// Nomad, and if so, continue the loop to try again.
		if revertDep.ID == dep.ID {
			log.Debug().Msgf("levant/auto_revert: auto-revert deployment not triggered for job %s, rechecking", dep.JobID)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Info().Msgf("levant/auto_revert: beginning deployment watcher for job %s", dep.JobID)
		success := l.deploymentWatcher(revertDep.ID)

		if success {
			log.Info().Msgf("levant/auto_revert: auto-revert of job %s was successful", dep.JobID)
			break
		}

		log.Error().Msgf("levant/auto_revert: auto-revert of job %s failed; POTENTIAL OUTAGE SITUATION", dep.JobID)
		l.checkFailedDeployment(&revertDep.ID)
		break
	}

	// At this point we have not been able to get the latest deploymentID that
	// is different from the original so we can't perform auto-revert checking.
	if i == 5 {
		log.Error().Msgf("levant/auto_revert: unable to check auto-revert of job %s", dep.JobID)
	}
}

// checkAutoRevert inspects a Nomad deployment to determine if any TashGroups
// have been auto-reverted.
func (l *levantDeployment) checkAutoRevert(dep *nomad.Deployment) {

	var revert bool

	// Identify whether any of the TaskGroups are enabled for auto-revert and have
	// therefore caused the job to enter a deployment to revert to a stable
	// version.
	for _, v := range dep.TaskGroups {
		if v.AutoRevert {
			revert = true
			break
		}
	}

	if revert {
		log.Info().Msgf("levant/auto_revert: job %v has entered auto-revert state; launching auto-revert checker",
			dep.JobID)

		// Run the levant autoRevert function.
		l.autoRevert(dep)
	} else {
		log.Info().Msgf("levant/auto_revert: job %v is not in auto-revert; POTENTIAL OUTAGE SITUATION", dep.JobID)
	}
}
