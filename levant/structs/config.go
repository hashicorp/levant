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

	// AllowStale sets consistency level for nomad query - https://www.nomadproject.io/api/index.html#consistency-modes
	AllowStale bool

	// Canary enables canary autopromote and is the value in seconds to wait
	// until attempting to perfrom autopromote.
	Canary int

	// ForceBatch is a boolean flag that can be used to force a run of a periodic
	// job upon registration.
	ForceBatch bool

	// ForceCount is a boolean flag that can be used to ignore running job counts
	// and force the count based on the rendered job file.
	ForceCount bool

	// IgnoreNoChanges is used to allow operators to force Levant to exit cleanly
	// even if there are no changes found during the plan.
	IgnoreNoChanges bool

	// Job represents the Nomad Job definition that will be deployed.
	Job *nomad.Job

	// LogLevel is the level at which Levant will log.
	LogLevel string

	// LogFormat is the format Levant will use for logging.
	LogFormat string

	// TemplateFile is the job specification template which will be rendered
	// before being deployed to the cluster.
	TemplateFile string

	// VariableFiles contains the variables which will be substituted into the
	// templateFile before deployment.
	VariableFiles []string
}
