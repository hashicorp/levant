package levant

import (
	nomad "github.com/hashicorp/nomad/api"
	"github.com/jrasell/levant/logging"
)

func (l *levantDeployment) autoRevert(jobID *string) {

	dep, _, err := l.nomad.Jobs().LatestDeployment(*jobID, nil)
	if err != nil {
		logging.Error("levant/auto_revert: unable to query latest deployment of job %s", *jobID)
		return
	}

	logging.Info("levant/auto_revert: beginning deployment watcher for job %s", *jobID)
	success := l.deploymentWatcher(dep.ID)

	if success {
		logging.Info("levant/auto_revert: auto-revert of job %s was successful", *jobID)
	} else {
		logging.Error("levant/auto_revert: auto-revert of job %s failed; POTENTIAL OUTAGE SITUATION", *jobID)
		l.checkFailedDeployment(&dep.ID)
	}
}

// checkAutoRevert inspects a Nomad deployment to determine if any TashGroups
// have been auto-reverted.
func (l *levantDeployment) checkAutoRevert(dep *nomad.Deployment) {

	var revert bool

	// Identify whether any of the TashGroups are enabled for auto-revert and have
	// therefore caused the job to enter a deployment to revert to a stable
	// version.
	for _, v := range dep.TaskGroups {
		if v.AutoRevert {
			revert = true
		}
	}

	if revert {
		logging.Info("levant/auto_revert: job %v has entered auto-revert state; launching auto-revert checker",
			dep.JobID)

		// Run the levant autoRevert function.
		l.autoRevert(&dep.JobID)
	} else {
		logging.Info("levant/auto_revert: job %v is not in auto-revert; POTENTIAL OUTAGE SITUATION", dep.JobID)
	}
}
