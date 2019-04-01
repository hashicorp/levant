package acctest

import (
	"fmt"

	"github.com/jrasell/levant/levant"
	"github.com/jrasell/levant/levant/structs"
	"github.com/jrasell/levant/template"
)

// DeployTestStepRunner implements TestStepRunner to execute a levant deployment
type DeployTestStepRunner struct {
	FixtureName string

	Canary      int
	ForceBatch  bool
	ForceCounts bool
}

// Run renders the job fixture and triggers a deployment
func (c DeployTestStepRunner) Run(s *TestState) error {
	vars := map[string]string{
		"job_name": s.JobName,
	}
	job, err := template.RenderJob("fixtures/"+c.FixtureName, []string{}, "", &vars)
	if err != nil {
		return fmt.Errorf("error rendering template: %s", err)
	}

	cfg := &levant.DeployConfig{
		Deploy: &structs.DeployConfig{
			Canary: c.Canary,
		},
		Client: &structs.ClientConfig{},
		Template: &structs.TemplateConfig{
			Job: job,
		},
	}

	if !levant.TriggerDeployment(cfg, nil) {
		return fmt.Errorf("deployment failed")
	}

	return nil
}
