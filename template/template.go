package template

import (
	"text/template"
)

const (
	terraformVarExtension = ".tf"
	yamlVarExtension      = ".yaml"
	ymlVarExtension       = ".yml"
)

// newTemplate returns an empty template with default options set
func newTemplate() *template.Template {
	return template.New("jobTemplate").Delims("[[", "]]").Option("missingkey=error")
}
