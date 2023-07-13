// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"github.com/rs/zerolog/log"
)

// VariableMerge merges the passed file variables with the flag variabes to
// provide a single set of variables. The flagVars will always prevale over file
// variables.
func VariableMerge(fileVars, flagVars *map[string]interface{}) map[string]interface{} {

	out := make(map[string]interface{})

	for k, v := range *flagVars {
		log.Info().Msgf("helper/variable: using command line variable with key %s and value %s", k, v)
		out[k] = v
	}

	for k, v := range *fileVars {
		if _, ok := out[k]; ok {
			log.Debug().Msgf("helper/variable: variable from file with key %s and value %s overridden by CLI var",
				k, v)
			continue
		}
		log.Info().Msgf("helper/variable: using variable with key %s and value %v from file", k, v)
		out[k] = v
	}

	return out
}
