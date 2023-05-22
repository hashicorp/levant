// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package levant

import (
	"fmt"
	"os"

	"github.com/hashicorp/levant/client"
	"github.com/hashicorp/levant/levant/structs"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog/log"
)

const (
	diffTypeAdded  = "Added"
	diffTypeEdited = "Edited"
	diffTypeNone   = "None"
)

type levantPlan struct {
	nomad  *nomad.Client
	config *PlanConfig

	options *nomad.WriteOptions
}

// PlanConfig is the set of config structs required to run a Levant plan.
type PlanConfig struct {
	Client   *structs.ClientConfig
	Plan     *structs.PlanConfig
	Template *structs.TemplateConfig
}

func newPlan(config *PlanConfig) (*levantPlan, error) {

	var err error

	plan := &levantPlan{}
	plan.config = config

	plan.nomad, err = client.NewNomadClient(config.Client.Addr)

	plan.options = setWriteOptions(plan.config.Template)

	if err != nil {
		return nil, err
	}
	return plan, nil
}

func setWriteOptions(template *structs.TemplateConfig) *nomad.WriteOptions {
	options := &nomad.WriteOptions{}

	if template.Job.Namespace != nil {
		options.Namespace = *template.Job.Namespace
	}
	if os.Getenv("NOMAD_NAMESPACE") != "" {
		log.Info().Msgf("levant/plan: using namespace from env-var: %s", os.Getenv("NOMAD_NAMESPACE"))
		options.Namespace = os.Getenv("NOMAD_NAMESPACE")
	}
	if template.Job.Region != nil {
		options.Region = *template.Job.Region
	}
	if os.Getenv("NOMAD_REGION") != "" {
		log.Info().Msgf("levant/plan: using region from env-var: %s", os.Getenv("NOMAD_REGION"))
		options.Namespace = os.Getenv("NOMAD_REGION")
	}
	return options
}

func setQueryOptions(wopt *nomad.WriteOptions) *nomad.QueryOptions {
	qopt := &nomad.QueryOptions{}
	qopt.Namespace = wopt.Namespace
	qopt.Region = wopt.Region
	return qopt
}

// TriggerPlan initiates a Levant plan run.
func TriggerPlan(config *PlanConfig) (bool, bool) {

	lp, err := newPlan(config)
	if err != nil {
		log.Error().Err(err).Msg("levant/plan: unable to setup Levant plan")
		return false, false
	}

	changes, err := lp.plan()
	if err != nil {
		log.Error().Err(err).Msg("levant/plan: error when running plan")
		return false, changes
	}

	if !changes && lp.config.Plan.IgnoreNoChanges {
		log.Info().Msg("levant/plan: no changes found in job but ignore-no-changes flag set to true")
	} else if !changes && !lp.config.Plan.IgnoreNoChanges {
		log.Info().Msg("levant/plan: no changes found in job")
		return false, changes
	}

	return true, changes
}

// plan is the entry point into running the Levant plan function which logs all
// changes anticipated by Nomad of the upcoming job registration. If there are
// no planned changes here, return false to indicate we should stop the process.
func (lp *levantPlan) plan() (bool, error) {

	log.Debug().Msg("levant/plan: triggering Nomad plan")

	// Run a plan using the rendered job.
	resp, _, err := lp.nomad.Jobs().Plan(lp.config.Template.Job, true, lp.options)
	if err != nil {
		log.Error().Err(err).Msg("levant/plan: unable to run a job plan")
		return false, err
	}

	switch resp.Diff.Type {

	// If the job is new, then don't print the entire diff but just log that it
	// is a new registration.
	case diffTypeAdded:
		log.Info().Msg("levant/plan: job is a new addition to the cluster")
		return true, nil

		// If there are no changes, log the message so the user can see this and
		// exit the deployment.
	case diffTypeNone:
		log.Info().Msg("levant/plan: no changes detected for job")
		return false, nil

		// If there are changes, run the planDiff function which is responsible for
		// iterating through the plan and logging all the planned changes.
	case diffTypeEdited:
		planDiff(resp.Diff)
	}

	return true, nil
}

func planDiff(plan *nomad.JobDiff) {

	// Iterate through each TaskGroup.
	for _, tg := range plan.TaskGroups {
		if tg.Type != diffTypeEdited {
			continue
		}
		for _, tgo := range tg.Objects {
			recurseObjDiff(tg.Name, "", tgo)
		}

		// Iterate through each Task.
		for _, t := range tg.Tasks {
			if t.Type != diffTypeEdited {
				continue
			}
			if len(t.Objects) == 0 {
				return
			}
			for _, o := range t.Objects {
				recurseObjDiff(tg.Name, t.Name, o)
			}
		}
	}
}

func recurseObjDiff(g, t string, objDiff *nomad.ObjectDiff) {

	// If we have reached the end of the object tree, and have an edited type
	// with field information then we can interate on the fields to find those
	// which have changed.
	if len(objDiff.Objects) == 0 && len(objDiff.Fields) > 0 && objDiff.Type == diffTypeEdited {
		for _, f := range objDiff.Fields {
			if f.Type != diffTypeEdited {
				continue
			}
			logDiffObj(g, t, objDiff.Name, f.Name, f.Old, f.New)
			continue
		}

	} else {
		// Continue to interate through the object diff objects until such time
		// the above is triggered.
		for _, o := range objDiff.Objects {
			recurseObjDiff(g, t, o)
		}
	}
}

// logDiffObj is a helper function so Levant can log the most accurate and
// useful plan output messages.
func logDiffObj(g, t, objName, fName, fOld, fNew string) {

	var lStart, l string

	// We will always have at least this information to log.
	lEnd := fmt.Sprintf("plan indicates change of %s:%s from %s to %s",
		objName, fName, fOld, fNew)

	// If we have been passed a group name, use this to start the log line.
	if g != "" {
		lStart = fmt.Sprintf("group %s ", g)
	}

	// If we have been passed a task name, append this to the group name.
	if t != "" {
		lStart = lStart + fmt.Sprintf("and task %s ", t)
	}

	// Build the final log message.
	if lStart != "" {
		l = lStart + lEnd
	} else {
		l = lEnd
	}

	log.Info().Msgf("levant/plan: %s", l)
}
