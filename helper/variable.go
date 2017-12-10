package helper

import "github.com/jrasell/levant/logging"

// VariableMerge merges the passed file variables with the flag variabes to
// provide a single set of variables. The flagVars will always prevale over file
// variables.
func VariableMerge(fileVars *map[string]interface{}, flagVars *map[string]string) map[string]interface{} {

	out := make(map[string]interface{})

	for k, v := range *flagVars {
		logging.Info("helper/variable: using command line variable with key %s and value %s", k, v)
		out[k] = v
	}

	for k, v := range *fileVars {
		if _, ok := out[k]; ok {
			logging.Debug("helper/variable: variable from file with key %s and value %s overridden by CLI var",
				k, v)
			continue
		}
		logging.Info("helper/variable: using variable with key %s and value %v from file", k, v)
		out[k] = v
	}

	return out
}
