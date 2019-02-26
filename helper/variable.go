package helper

import (
	"github.com/imdario/mergo"
)

// VariableMerge merges the passed file variables with the flag variabes to
// provide a single set of variables. The flagVars will always prevale over file
// variables.
func VariableMerge(fileVars *map[string]interface{}, flagVars *map[string]interface{}) map[string]interface{} {
	if err := mergo.Merge(&*fileVars, *flagVars, mergo.WithOverride); err != nil {
		panic(err)
	}
	return *fileVars
}
