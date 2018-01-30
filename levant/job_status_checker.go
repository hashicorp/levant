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
func (c *nomadClient) checkJobStatus(jobName *string) bool {

	logging.Info("levant/job_status_checker: running job status checker for %s", *jobName)

	// Initialiaze our WaitIndex
	var wi uint64

	// Setup the Nomad QueryOptions to allow blocking query and a timeout.
	q := &nomad.QueryOptions{WaitIndex: wi}
	timeout := time.Tick(time.Minute * 5)

	for {

		job, meta, err := c.nomad.Jobs().Info(*jobName, q)
		if err != nil {
			logging.Error("levant/job_status_checker: unable to query batch job %s: %v", *jobName, err)
			return false
		}

		// If the LastIndex is not greater than our stored LastChangeIndex, we don't
		// need to do anything.
		if meta.LastIndex <= wi {
			continue
		}

		if *job.Status == nomadStructs.JobStatusRunning {
			logging.Info("levant/job_status_checker: job %s has status %s", *jobName, *job.Status)
			return true
		}

		select {
		case <-timeout:
			logging.Error("levant/job_status_checker: timeout reached while verifying the status of job %s",
				*jobName)
			return false
		default:
			logging.Debug("levant/job_status_checker: job %s currently has status %s", *jobName, *job.Status)
			q.WaitIndex = meta.LastIndex
			continue
		}
	}
}
