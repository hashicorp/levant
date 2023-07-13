// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"reflect"
	"testing"
)

func TestHelper_VariableMerge(t *testing.T) {

	flagVars := make(map[string]interface{})
	flagVars["job_name"] = "levantExample"
	flagVars["datacentre"] = "dc13"

	fileVars := make(map[string]interface{})
	fileVars["job_name"] = "levantExampleOverride"
	fileVars["CPU_MHz"] = 500

	expected := make(map[string]interface{})
	expected["job_name"] = "levantExample"
	expected["datacentre"] = "dc13"
	expected["CPU_MHz"] = 500

	res := VariableMerge(&fileVars, &flagVars)

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("expected \n%#v\n\n, got \n\n%#v\n\n", expected, res)
	}
}
