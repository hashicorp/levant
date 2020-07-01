package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/jrasell/levant/client"
	"github.com/jrasell/levant/helper"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/hashicorp/terraform/config"
)

// RenderJob takes in a template and variables performing a render of the
// template followed by Nomad jobspec parse.
func RenderJob(templateFile string, variableFiles []string, nomadAddr string, consulAddr string, flagVars *map[string]string) (job *nomad.Job, err error) {
	var tpl *bytes.Buffer
	tpl, err = RenderTemplate(templateFile, variableFiles, consulAddr, flagVars)
	if err != nil {
		return
	}

	var n *nomad.Client
	n, err = client.NewNomadClient(nomadAddr)
	if err != nil {
		return
	}

	job, err = n.Jobs().ParseHCL(tpl.String(), false)
	return
}

// RenderTemplate is the main entry point to render the template based on the
// passed variables file.
func RenderTemplate(templateFile string, variableFiles []string, addr string, flagVars *map[string]string) (tpl *bytes.Buffer, err error) {

	t := &tmpl{}
	t.flagVariables = flagVars
	t.jobTemplateFile = templateFile
	t.variableFiles = variableFiles

	c, err := client.NewConsulClient(addr)
	if err != nil {
		return
	}

	t.consulClient = c

	if len(variableFiles) == 0 {
		log.Debug().Msgf("template/render: no variable file passed, trying defaults")
		defaultVarFile := helper.GetDefaultVarFile()
		if defaultVarFile != "" {
			t.variableFiles = make([]string, 1)
			t.variableFiles[0] = defaultVarFile
			log.Debug().Msgf("template/render: found default variable file, using %s", t.variableFiles[0])
		}
	}

	mergedVariables := make(map[string]interface{})
	for _, variableFile := range variableFiles {
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

func (t *tmpl) parseTFVars(variableFile string) (variables map[string]interface{}, err error) {

	c := &config.Config{}
	if c, err = config.LoadFile(variableFile); err != nil {
		return
	}

	variables = make(map[string]interface{})
	for _, variable := range c.Variables {
		variables[variable.Name] = variable.Default
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
