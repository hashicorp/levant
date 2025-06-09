// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/rs/zerolog/log"
)

// JobJSON is used to unmarshal pre-rendered/parsed jobspec JSON
type JobJSON struct {
	Job nomad.Job `json:"Job"`
}

// GetDefaultTmplFile checks the current working directory for *.nomad files.
// If only 1 is found we return the match.
func GetDefaultTmplFile() (templateFile string) {
	if matches, _ := filepath.Glob("*.nomad"); matches != nil {
		if len(matches) == 1 {
			templateFile = matches[0]
			log.Debug().Msgf("helper/files: using templatefile `%v`", templateFile)
			return templateFile
		}
	}
	return ""
}

// GetDefaultVarFile checks the current working directory for levant.(yaml|yml|tf) files.
// The first match is returned.
func GetDefaultVarFile() (varFile string) {
	if _, err := os.Stat("levant.yaml"); !os.IsNotExist(err) {
		log.Debug().Msg("helper/files: using default var-file `levant.yaml`")
		return "levant.yaml"
	}
	if _, err := os.Stat("levant.yml"); !os.IsNotExist(err) {
		log.Debug().Msg("helper/files: using default var-file `levant.yml`")
		return "levant.yml"
	}
	if _, err := os.Stat("levant.json"); !os.IsNotExist(err) {
		log.Debug().Msg("helper/files: using default var-file `levant.json`")
		return "levant.json"
	}
	if _, err := os.Stat("levant.tf"); !os.IsNotExist(err) {
		log.Debug().Msg("helper/files: using default var-file `levant.tf`")
		return "levant.tf"
	}
	log.Debug().Msg("helper/files: no default var-file found")
	return ""
}

// GetJobspecFromBytes converts JSON passed as bytes to a nomad job the same
// as levant's RenderTemplate would return.
func GetJobspecFromBytes(src []byte) (job *nomad.Job, err error) {
	var jobspec JobJSON

	err = json.Unmarshal(src, &jobspec)
	if err != nil {
		err = fmt.Errorf("helper/files: error parsing JSON: %w", err)
	}

	return &jobspec.Job, err
}

// GetJobspecFromFile converts a JSON file to a nomad job the same as levant's
// RenderTemplate would return.
func GetJobspecFromFile(jobFile string) (job *nomad.Job, err error) {
	src, err := ioutil.ReadFile(jobFile)
	if err != nil {
		return nil, err
	}

	return GetJobspecFromBytes(src)
}

// IsPipedInput determines if there is piped input to stdin.
func IsPipedInput() (isPiped bool, err error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}

	return info.Mode()&os.ModeCharDevice == 0, err
}

// getJobspecFromIOReader reads bytes from a Reader and returns a nomad Job.
// Intended for os.Stdin but can be tested with any io.Reader.
// JSON must be valid and conform to the nomad.api.Job struct.
func GetJobspecFromIOReader(r io.Reader) (job *nomad.Job, err error) {
	var runes []rune
	reader := bufio.NewReader(r)

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		runes = append(runes, input)
	}

	return GetJobspecFromBytes([]byte(string(runes)))
}
