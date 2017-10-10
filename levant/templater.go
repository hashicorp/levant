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
	ext := path.Ext(variableFile)

	src, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return
	}

	switch ext {
	case terraformVarExtention:
		logging.Debug("levant/templater: detected .tf variable file extension")
		tpl, err = renderTFTemplte(string(src), variableFile, flagVars)
	case yamlVarExtension, ymlVarExtension:
		logging.Debug("levant/templater: detected .yaml or .yml variable file extension")
		tpl, err = renderYAMLVarsTemplate(string(src), variableFile, flagVars)
	case "":
		if len(*flagVars) == 0 {
			tpl = bytes.NewBuffer(src)
		}

		logging.Debug("levant/templater: variable file not passed, using any passed CLI variables")
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
