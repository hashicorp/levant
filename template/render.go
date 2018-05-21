package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/jrasell/levant/client"
	"github.com/jrasell/levant/helper"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/hashicorp/terraform/config"
)

// RenderJob takes in a template and variables performing a render of the
// template followed by Nomad jobspec parse.
func RenderJob(templateFile, variableFile, addr string, flagVars *map[string]string) (job *nomad.Job, err error) {
	var tpl *bytes.Buffer
	tpl, err = RenderTemplate(templateFile, variableFile, addr, flagVars)
	if err != nil {
		return
	}

	job, err = jobspec.Parse(tpl)
	return
}

// RenderTemplate is the main entry point to render the template based on the
// passed variables file.
func RenderTemplate(templateFile, variableFile, addr string, flagVars *map[string]string) (tpl *bytes.Buffer, err error) {

	t := &tmpl{}
	t.flagVariables = flagVars
	t.jobTemplateFile = templateFile
	t.variableFile = variableFile

	c, err := client.NewConsulClient(addr)
	if err != nil {
		return
	}

	t.consulClient = c

	if variableFile == "" {
		log.Debug().Msgf("template/render: no variable file passed, trying defaults")
		if t.variableFile = helper.GetDefaultVarFile(); t.variableFile != "" {
			log.Debug().Msgf("template/render: found default variable file, using %s", t.variableFile)
		}
	}

	// Process the variable file extension and log DEBUG so the template can be
	// correctly rendered.
	var ext string
	if ext = path.Ext(t.variableFile); ext != "" {
		log.Debug().Msgf("template/render: variable file extension %s detected", ext)
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

	switch ext {
	case terraformVarExtension:
		tpl, err = t.renderTF(string(src))

	case yamlVarExtension, ymlVarExtension:
		// Run the render using a YAML variable file.
		tpl, err = t.renderYAMLVars(string(src))

	case "":
		// No variables file passed; render using any passed CLI variables.
		log.Debug().Msgf("template/render: variable file not passed")
		tpl, err = t.readJobFile(string(src))

	default:
		err = fmt.Errorf("variables file extension %v not supported", ext)
	}

	return
}

func (t *tmpl) renderTF(src string) (tpl *bytes.Buffer, err error) {

	c := &config.Config{}

	// Setup our variables map and template
	variables := make(map[string]interface{})
	tpl = &bytes.Buffer{}

	// Load the variables file
	if c, err = config.LoadFile(t.variableFile); err != nil {
		return
	}

	// Populate our map with the parsed variables
	for _, variable := range c.Variables {
		variables[variable.Name] = variable.Default
	}

	// Merge variables passed on the CLI with those passed through a variables
	// file.
	mergedVars := helper.VariableMerge(&variables, t.flagVariables)

	// Setup the template file for rendering
	tmpl := t.newTemplate()

	if tmpl, err = tmpl.Parse(src); err != nil {
		return
	}

	if err = tmpl.Execute(tpl, mergedVars); err != nil {
		return
	}
	return tpl, nil
}

func (t *tmpl) renderYAMLVars(src string) (tpl *bytes.Buffer, err error) {

	// Setup our variables map and template
	variables := make(map[string]interface{})
	tpl = &bytes.Buffer{}

	yamlFile, err := ioutil.ReadFile(t.variableFile)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(yamlFile, &variables); err != nil {
		return
	}

	// Merge variables passed on the CLI with those passed through a variables
	// file.
	mergedVars := helper.VariableMerge(&variables, t.flagVariables)

	// Setup the template file for rendering
	tmpl := t.newTemplate()
	if tmpl, err = tmpl.Parse(string(src)); err != nil {
		return
	}

	if err = tmpl.Execute(tpl, mergedVars); err != nil {
		return
	}

	return tpl, nil
}

func (t *tmpl) readJobFile(src string) (tpl *bytes.Buffer, err error) {

	tpl = &bytes.Buffer{}

	// Setup the template file for rendering
	tmpl := t.newTemplate()
	if tmpl, err = tmpl.Parse(src); err != nil {
		return
	}

	if err = tmpl.Execute(tpl, t.flagVariables); err != nil {
		return
	}

	return tpl, nil
}
