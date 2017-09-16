package helper

// VariableMerge merges the passed file variables with the flag variabes to
// provide a single set of variables. The flagVars will always prevale over file
// variables.
func VariableMerge(fileVars *map[string]interface{}, flagVars *map[string]string) map[string]interface{} {

	out := make(map[string]interface{})

	for k, v := range *flagVars {
		out[k] = v
	}

	for k, v := range *fileVars {
		if _, ok := out[k]; ok {
			continue
		}
		out[k] = v
	}

	return out
}
