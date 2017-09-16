package levant

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/jrasell/levant/helper"
	"github.com/jrasell/levant/logging"
	yaml "gopkg.in/yaml.v2"

	"github.com/hashicorp/nomad/jobspec"
	"github.com/hashicorp/terraform/config"
)

const (
	terraformVarExtention = "tf"
	yamlVarExtension      = "yaml"
	ymlVarExtension       = "yml"
)

// RenderTemplate is the main entry point to render the template based on the
// passed variables file.
func RenderTemplate(templateFile, variableFile string, flagVars *map[string]string) (job *nomad.Job, err error) {

	var tpl bytes.Buffer

	s := strings.Split(variableFile, ".")
	ext := s[len(s)-1]

	switch ext {
	case terraformVarExtention:
		logging.Debug("levant/templater: detected .tf variable file extension")
		tpl, err = renderTFTemplte(templateFile, variableFile, flagVars)
	case yamlVarExtension, ymlVarExtension:
		logging.Debug("levant/templater: detected .yaml or .yml variable file extension")
		tpl, err = renderYAMLVarsTemplate(templateFile, variableFile, flagVars)
	case "":
		logging.Debug("levant/templater: variable file not passed, using any passed CLI variables")
		tpl, err = readJobFile(templateFile, flagVars)
	default:
		err = fmt.Errorf("variables file extension %v not supported", ext)
	}

	if err != nil {
		return
	}

	// Parse the rendered template as a Nomad job.
	if job, err = jobspec.Parse(&tpl); err != nil {
		return
	}
	return
}

func renderTFTemplte(templateFile, variableFile string, flagVars *map[string]string) (tpl bytes.Buffer, err error) {

	c := &config.Config{}

	// Setup our variables map and template
	variables := make(map[string]interface{})

	// Load the variables file
	if c, err = config.LoadFile(variableFile); err != nil {
		return
	}

	// Populate our map with the parsed variables
	for _, variable := range c.Variables {
		variables[variable.Name] = variable.Default
	}

	// Merge variables passed on the CLI with those passed through a variables
	// file.
	mergedVars := helper.VariableMerge(&variables, flagVars)

	// Setup the template file for rendering
	t := template.New("jobTemplate")
	tempF, _ := ioutil.ReadFile(templateFile)
	if t, err = t.Parse(string(tempF)); err != nil {
		return
	}

	if err = t.Execute(&tpl, mergedVars); err != nil {
		return
	}
	return tpl, nil
}

func renderYAMLVarsTemplate(templateFile, variableFile string, flagVars *map[string]string) (tpl bytes.Buffer, err error) {

	// Setup our variables map and template
	variables := make(map[string]interface{})

	yamlFile, err := ioutil.ReadFile(variableFile)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(yamlFile, &variables); err != nil {
		return
	}

	// Merge variables passed on the CLI with those passed through a variables
	// file.
	mergedVars := helper.VariableMerge(&variables, flagVars)

	// Setup the template file for rendering
	t := template.New("jobTemplate")
	tempF, _ := ioutil.ReadFile(templateFile)
	if t, err = t.Parse(string(tempF)); err != nil {
		return
	}

	if err = t.Execute(&tpl, mergedVars); err != nil {
		return
	}

	return tpl, nil
}

func readJobFile(templateFile string, flagVars *map[string]string) (tpl bytes.Buffer, err error) {

	// Setup the template file for rendering
	t := template.New("jobTemplate")
	tempF, _ := ioutil.ReadFile(templateFile)
	if t, err = t.Parse(string(tempF)); err != nil {
		return
	}

	if err = t.Execute(&tpl, flagVars); err != nil {
		return
	}

	return tpl, nil
}
