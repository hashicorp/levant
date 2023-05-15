// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package structs

import nomad "github.com/hashicorp/nomad/api"

const (
	// JobIDContextField is the logging context feild added when interacting
	// with jobs.
	JobIDContextField = "job_id"

	// ScalingDirectionOut represents a scaling out event; adding to the total number.
	ScalingDirectionOut = "Out"

	// ScalingDirectionIn represents a scaling in event; removing from the total number.
	ScalingDirectionIn = "In"

	// ScalingDirectionTypeCount means the scale event will use a change by count.
	ScalingDirectionTypeCount = "Count"

	// ScalingDirectionTypePercent means the scale event will use a percentage of current change.
	ScalingDirectionTypePercent = "Percent"
)

// DeployConfig is the main struct used to configure and run a Levant deployment on
// a given target job.
type DeployConfig struct {
	// Canary enables canary autopromote and is the value in seconds to wait
	// until attempting to perform autopromote.
	Canary int

	// Force is a boolean flag that can be used to force a deployment
	// even though levant didn't detect any changes.
	Force bool

	// ForceBatch is a boolean flag that can be used to force a run of a periodic
	// job upon registration.
	ForceBatch bool

	// ForceCount is a boolean flag that can be used to ignore running job counts
	// and force the count based on the rendered job file.
	ForceCount bool

	// EnvVault is a boolean flag that can be used to enable reading the VAULT_TOKEN
	// from the enviromment.
	EnvVault bool

	// VaultToken is a string with the vault token.
	VaultToken string
}

// ClientConfig is the config struct which houses all the information needed to connect
// to the external services and endpoints.
type ClientConfig struct {
	// Addr is the Nomad API address to use for all calls and must include both
	// protocol and port.
	Addr string

	// ConsulAddr is the Consul API address to use for all calls.
	ConsulAddr string

	// AllowStale sets consistency level for nomad query
	// https://www.nomadproject.io/api/index.html#consistency-modes
	AllowStale bool
}

// PlanConfig contains any configuration options that are specific to running a
// Nomad plan.
type PlanConfig struct {
	// IgnoreNoChanges is used to allow operators to force Levant to exit cleanly
	// even if there are no changes found during the plan.
	IgnoreNoChanges bool
}

// TemplateConfig contains all the job templating configuration options including
// the rendered job.
type TemplateConfig struct {
	// Job represents the Nomad Job definition that will be deployed.
	Job *nomad.Job

	// TemplateFile is the job specification template which will be rendered
	// before being deployed to the cluster.
	TemplateFile string

	// VariableFiles contains the variables which will be substituted into the
	// templateFile before deployment.
	VariableFiles []string
}

// ScaleConfig contains all the scaling specific configuration options.
type ScaleConfig struct {
	// Count is the count by which the operator has asked to scale the Nomad job
	// and optional taskgroup by.
	Count int

	// Direction is the direction in which the scaling will take place and is
	// populated by consts.
	Direction string

	// DirectionType is an identifier on whether the operator has specified to
	// scale using a count increase or percentage.
	DirectionType string

	// JobID is the Nomad job which will be interacted with for scaling.
	JobID string

	// Percent is the percentage by which the operator has asked to scale the
	// Nomad job and optional taskgroup by.
	Percent int

	// TaskGroup is the Nomad job taskgroup which has been selected for scaling.
	TaskGroup string
}
