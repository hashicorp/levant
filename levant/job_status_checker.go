package levant

import (
	"time"

	nomad "github.com/hashicorp/nomad/api"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/jrasell/levant/logging"
)

// checkJobStatus checks the status of a job at least reaches a status of
// running. This is required as currently Nomad does not support deployments
// across all job types.
func (l *levantDeployment) checkJobStatus() bool {

	j := l.config.Job.Name
	logging.Info("levant/job_status_checker: running job status checker for %s", *j)

	// Initialiaze our WaitIndex
	var wi uint64

	// Setup the Nomad QueryOptions to allow blocking query and a timeout.
	q := &nomad.QueryOptions{WaitIndex: wi}
	timeout := time.Tick(time.Minute * 5)

	for {

		job, meta, err := l.nomad.Jobs().Info(*j, q)
		if err != nil {
			logging.Error("levant/job_status_checker: unable to query batch job %s: %v", *j, err)
			return false
		}

		// If the LastIndex is not greater than our stored LastChangeIndex, we don't
		// need to do anything.
		if meta.LastIndex <= wi {
			continue
		}

		if *job.Status == nomadStructs.JobStatusRunning {
			logging.Info("levant/job_status_checker: job %s has status %s", *j, *job.Status)
			return true
		}

		select {
		case <-timeout:
			logging.Error("levant/job_status_checker: timeout reached while verifying the status of job %s",
				*j)
			return false
		default:
			logging.Debug("levant/job_status_checker: job %s currently has status %s", *j, *job.Status)
			q.WaitIndex = meta.LastIndex
			continue
		}
	}
}
