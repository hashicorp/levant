package structs

import nomad "github.com/hashicorp/nomad/api"

// Config is the main struct used to configure and run a Levant deployment on
// a given target job.
type Config struct {
	// Addr is the Nomad API address to use for all calls and must include both
	// protocol and port.
	Addr string

	// Canary enables canary autopromote and is the value in seconds to wait
	// until attempting to perfrom autopromote.
	Canary int

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
