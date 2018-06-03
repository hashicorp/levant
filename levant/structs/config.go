package structs

import nomad "github.com/hashicorp/nomad/api"

const (
	// JobIDContextField is the logging context feild added when interacting
	// with jobs.
	JobIDContextField = "job_id"
)

// Config is the main struct used to configure and run a Levant deployment on
// a given target job.
type Config struct {
	// Addr is the Nomad API address to use for all calls and must include both
	// protocol and port.
	Addr string

	// Canary enables canary autopromote and is the value in seconds to wait
	// until attempting to perfrom autopromote.
	Canary int

	// ExitAfterAutoRevert is a boolean flag that determine whether Levant will exit (with non-zero exit code)
	// after a deployment auto-reverts to the previous stable job. This flag is useful if you want to be notified
	// in case auto-revert happens (e.g. to fail your CD pipeline).
	ExitAfterAutoRevert bool

	// ForceBatch is a boolean flag that can be used to force a run of a periodic
	// job upon registration.
	ForceBatch bool

	// ForceCount is a boolean flag that can be used to ignore running job counts
	// and force the count based on the rendered job file.
	ForceCount bool

	// Job represents the Nomad Job definition that will be deployed.
	Job *nomad.Job

	// LogLevel is the level at which Levant will log.
	LogLevel string

	// LogFormat is the format Levant will use for logging.
	LogFormat string

	// TemplateFile is the job specification template which will be rendered
	// before being deployed to the cluster.
	TemplateFile string

	// VaiableFile contains the variables which will be substituted into the
	// templateFile before deployment.
	VaiableFile string
}
