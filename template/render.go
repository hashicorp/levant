// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/hashicorp/levant/client"
	"github.com/hashicorp/levant/helper"
	nomad "github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/hcl2shim"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
)

// RenderJob takes in a template and variables performing a render of the
// template followed by Nomad jobspec parse.
func RenderJob(templateFile string, variableFiles []string, addr string, flagVars *map[string]interface{}) (job *nomad.Job, err error) {
	var tpl *bytes.Buffer
	tpl, err = RenderTemplate(templateFile, variableFiles, addr, flagVars)
	if err != nil {
		return
	}

	return jobspec.Parse(tpl)
}

// RenderTemplate is the main entry point to render the template based on the
// passed variables file.
func RenderTemplate(templateFile string, variableFiles []string, addr string, flagVars *map[string]interface{}) (tpl *bytes.Buffer, err error) {

	t := &tmpl{}
	t.flagVariables = flagVars
	t.jobTemplateFile = templateFile
	t.variableFiles = variableFiles

	c, err := client.NewConsulClient(addr)
	if err != nil {
		return
	}

	t.consulClient = c

	if len(t.variableFiles) == 0 {
		log.Debug().Msgf("template/render: no variable file passed, trying defaults")
		if defaultVarFile := helper.GetDefaultVarFile(); defaultVarFile != "" {
			t.variableFiles = []string{defaultVarFile}
			log.Debug().Msgf("template/render: found default variable file, using %s", defaultVarFile)
		}
	}

	mergedVariables := make(map[string]interface{})
	for _, variableFile := range t.variableFiles {
		// Process the variable file extension and log DEBUG so the template can be
		// correctly rendered.
		var ext string
		if ext = path.Ext(variableFile); ext != "" {
			log.Debug().Msgf("template/render: variable file extension %s detected", ext)
		}

		var variables map[string]interface{}
		switch ext {
		case terraformVarExtension:
			variables, err = t.parseTFVars(variableFile)
		case yamlVarExtension, ymlVarExtension:
			variables, err = t.parseYAMLVars(variableFile)
		case jsonVarExtension:
			variables, err = t.parseJSONVars(variableFile)
		default:
			err = fmt.Errorf("variables file extension %v not supported", ext)
		}

		if err != nil {
			return
		}
		for k, v := range variables {
			mergedVariables[k] = v
		}
	}

	src, err := ioutil.ReadFile(t.jobTemplateFile)
	if err != nil {
		return
	}

	// If no command line variables are passed; log this as DEBUG to provide much
	// greater feedback.
	if len(*t.flagVariables) == 0 {
		log.Debug().Msgf("template/render: no command line variables passed")
	}

	tpl, err = t.renderTemplate(string(src), mergedVariables)

	return
}

func (t *tmpl) parseJSONVars(variableFile string) (variables map[string]interface{}, err error) {

	jsonFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return
	}

	variables = make(map[string]interface{})
	if err = json.Unmarshal(jsonFile, &variables); err != nil {
		return
	}

	return variables, nil
}

func (t *tmpl) parseTFVars(variableFile string) (map[string]interface{}, error) {

	tfParser := configs.NewParser(nil)
	loadedFile, loadDiags := tfParser.LoadConfigFile(variableFile)
	if loadDiags != nil && loadDiags.HasErrors() {
		return nil, loadDiags
	}
	if loadedFile == nil {
		return nil, fmt.Errorf("hcl returned nil file")
	}

	variables := make(map[string]interface{})
	for _, variable := range loadedFile.Variables {
		variables[variable.Name] = hcl2shim.ConfigValueFromHCL2(variable.Default)
	}
	return variables, nil
}

func (t *tmpl) parseYAMLVars(variableFile string) (variables map[string]interface{}, err error) {

	yamlFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return
	}

	variables = make(map[string]interface{})
	if err = yaml.Unmarshal(yamlFile, &variables); err != nil {
		return
	}
	return variables, nil
}

func (t *tmpl) renderTemplate(src string, variables map[string]interface{}) (tpl *bytes.Buffer, err error) {

	tpl = &bytes.Buffer{}

	// Setup the template file for rendering
	tmpl := t.newTemplate()
	if tmpl, err = tmpl.Parse(src); err != nil {
		return
	}

	if variables != nil {
		// Merge variables passed on the CLI with those passed through a variables file.
		err = tmpl.Execute(tpl, helper.VariableMerge(&variables, t.flagVariables))
	} else {
		// No variables file passed; render using any passed CLI variables.
		log.Debug().Msgf("template/render: variable file not passed")
		err = tmpl.Execute(tpl, t.flagVariables)
	}

	return tpl, err
}
