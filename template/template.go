package template

import (
	"text/template"

	consul "github.com/hashicorp/consul/api"
)

// tmpl provides everything needed to fully render and job template using
// inbuilt functions.
type tmpl struct {
	consulClient    *consul.Client
	flagVariables   *map[string]interface{}
	jobTemplateFile string
	variableFiles   []string

	// callStack contains the current stack of template calls. Not threadsafe.
	callStack []string
}

const (
	jsonVarExtension      = ".json"
	terraformVarExtension = ".tf"
	yamlVarExtension      = ".yaml"
	ymlVarExtension       = ".yml"
	rightDelim            = "]]"
	leftDelim             = "[["
)

// newTemplate returns an empty template with default options set
func (t *tmpl) newTemplate() *template.Template {
	tmpl := template.New("jobTemplate")
	tmpl.Delims(leftDelim, rightDelim)
	tmpl.Option("missingkey=zero")
	tmpl.Funcs(funcMap(t))
	return tmpl
}

// pushCall pushes a template path to the call stack.
func (t *tmpl) pushCall(tmplPath string) {
	t.callStack = append(t.callStack, tmplPath)
}

// popCall pops & returns the current top template path from the call stack.
// The bool return value is true iff there was an item to return in the stack.
func (t *tmpl) popCall() (string, bool) {
	l := len(t.callStack)
	if l == 0 {
		return "", false
	}
	var top string
	t.callStack, top = t.callStack[:l-1], t.callStack[l-1]
	return top, true
}

// callStackContains returns true iff tmplPath was pushed but not yet popped.
func (t *tmpl) callStackContains(tmplPath string) bool {
	for _, call := range t.callStack {
		if tmplPath == call {
			return true
		}
	}
	return false
}
