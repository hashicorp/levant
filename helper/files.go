package helper

import (
	"os"
	"path/filepath"

	"github.com/jrasell/levant/logging"
)

// GetDefaultTmplFile checks the current working directory for *.nomad files.
// If only 1 is found we return the match.
func GetDefaultTmplFile() (templateFile string) {
	if matches, _ := filepath.Glob("*.nomad"); matches != nil {
		if len(matches) == 1 {
			templateFile = matches[0]
			logging.Debug("helper/files: using templatefile `%v`", templateFile)
			return templateFile
		}
	}
	return ""
}

// GetDefaultVarFile checks the current working directory for levant.(yaml|yml|tf) files.
// The first match is returned.
func GetDefaultVarFile() (varFile string) {
	if _, err := os.Stat("levant.yaml"); !os.IsNotExist(err) {
		logging.Debug("helper/files: using default var-file `levant.yaml`")
		return "levant.yaml"
	}
	if _, err := os.Stat("levant.yml"); !os.IsNotExist(err) {
		logging.Debug("helper/files: using default var-file `levant.yml`")
		return "levant.yml"
	}
	if _, err := os.Stat("levant.tf"); !os.IsNotExist(err) {
		logging.Debug("helper/files: using default var-file `levant.tf`")
		return "levant.tf"
	}
	logging.Debug("helper/files: no default var-file found")
	return ""
}
