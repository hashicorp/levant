package levant

import (
	"testing"

	"github.com/hashicorp/levant/levant/structs"
	nomad "github.com/hashicorp/nomad/api"
)

func TestHasUpdateStanza(t *testing.T) {
	ld1 := levantDeployment{config: &DeployConfig{Template: &structs.TemplateConfig{Job: &nomad.Job{
		Update: nil}}}}
	ld2 := levantDeployment{config: &DeployConfig{Template: &structs.TemplateConfig{Job: &nomad.Job{
		Update: &nomad.UpdateStrategy{}}}}}
	ld3 := levantDeployment{config: &DeployConfig{Template: &structs.TemplateConfig{Job: &nomad.Job{
		Update: nil, TaskGroups: []*nomad.TaskGroup{{Update: nil}}}}}}
	ld4 := levantDeployment{config: &DeployConfig{Template: &structs.TemplateConfig{Job: &nomad.Job{
		Update: nil, TaskGroups: []*nomad.TaskGroup{{Update: &nomad.UpdateStrategy{}}}}}}}
	ld5 := levantDeployment{config: &DeployConfig{Template: &structs.TemplateConfig{Job: &nomad.Job{
		Update: nil, TaskGroups: []*nomad.TaskGroup{{Update: nil}, {Update: &nomad.UpdateStrategy{}}}}}}}

	cases := []struct {
		ld           levantDeployment
		expectedTrue bool
	}{
		{ld1, false},
		{ld2, true},
		{ld3, false},
		{ld4, true},
		{ld5, false},
	}

	for _, c := range cases {
		if c.ld.hasUpdateStanza() != c.expectedTrue {
			t.Fatalf("expected hasUpdate to be %t, but got %t", c.ld.hasUpdateStanza(), c.expectedTrue)
		}
	}
}
