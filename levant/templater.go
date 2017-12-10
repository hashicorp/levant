package levant

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"text/template"

	"github.com/jrasell/levant/helper"
	"github.com/jrasell/levant/logging"
	yaml "gopkg.in/yaml.v2"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/hashicorp/terraform/config"
)

const (
	terraformVarExtention = ".tf"
	yamlVarExtension      = ".yaml"
	ymlVarExtension       = ".yml"
)

// RenderJob takes in a template and variables performing a render of the
// template followed by Nomad jobspec parse.
func RenderJob(templateFile, variableFile string, flagVars *map[string]string) (job *nomad.Job, err error) {
	var tpl *bytes.Buffer
	tpl, err = RenderTemplate(templateFile, variableFile, flagVars)
	if err != nil {
		return
	}

	job, err = jobspec.Parse(tpl)
	return
}

// RenderTemplate is the main entry point to render the template based on the
// passed variables file.
func RenderTemplate(templateFile, variableFile string, flagVars *map[string]string) (tpl *bytes.Buffer, err error) {

	// Process the variable file extension and log DEBUG so the template can be
	// correctly rendered.
	ext := path.Ext(variableFile)
	if variableFile != "" {
		logging.Debug("levant/templater: variable file extension %s detected", ext)
	}

	src, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return
	}

	// If no command line variables are passed; log this as DEBUG to provide much
	// greater feedback.
	if len(*flagVars) == 0 {
		logging.Debug("levant/templater: no command line variables passed")
	}

	switch ext {
	case terraformVarExtention:
		// Run the render using variables formatted in Terraforms .tf extension.
		tpl, err = renderTFTemplte(string(src), variableFile, flagVars)

	case yamlVarExtension, ymlVarExtension:
		// Run the render using a YAML varaible file.
		tpl, err = renderYAMLVarsTemplate(string(src), variableFile, flagVars)

	case "":
		// No varibles file passed; render using any passed CLI variables.
		logging.Debug("levant/templater: variable file not passed")
		tpl, err = readJobFile(string(src), flagVars)

	default:
		err = fmt.Errorf("variables file extension %v not supported", ext)
	}

	return
}

func renderTFTemplte(src, variableFile string, flagVars *map[string]string) (tpl *bytes.Buffer, err error) {

	c := &config.Config{}

	// Setup our variables map and template
	variables := make(map[string]interface{})
	tpl = &bytes.Buffer{}

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
	t := newTemplate()

	if t, err = t.Parse(src); err != nil {
		return
	}

	if err = t.Execute(tpl, mergedVars); err != nil {
		return
	}
	return tpl, nil
}

func renderYAMLVarsTemplate(src, variableFile string, flagVars *map[string]string) (tpl *bytes.Buffer, err error) {

	// Setup our variables map and template
	variables := make(map[string]interface{})
	tpl = &bytes.Buffer{}

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
	t := newTemplate()
	if t, err = t.Parse(string(src)); err != nil {
		return
	}

	if err = t.Execute(tpl, mergedVars); err != nil {
		return
	}

	return tpl, nil
}

func readJobFile(src string, flagVars *map[string]string) (tpl *bytes.Buffer, err error) {

	tpl = &bytes.Buffer{}

	// Setup the template file for rendering
	t := newTemplate()
	if t, err = t.Parse(src); err != nil {
		return
	}

	if err = t.Execute(tpl, flagVars); err != nil {
		return
	}

	return tpl, nil
}

// newTemplate returns an empty template with default options set
func newTemplate() *template.Template {
	return template.New("jobTemplate").Delims("[[", "]]").Option("missingkey=error")
}
