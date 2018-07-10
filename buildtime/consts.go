package buildtime

// These variables are populated by govvv during build time to provide detailed
// version output information.
var (
	BuildDate  string
	GitCommit  string
	GitBranch  string
	GitState   string
	GitSummary string
	Version    string
)

const (
	// PROGNAME is the name of this application.
	PROGNAME = "Levant"
)
