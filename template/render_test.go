// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package template

import (
	"os"
	"testing"

	nomad "github.com/hashicorp/nomad/api"
)

const (
	testJobName           = "levantExample"
	testJobNameOverwrite  = "levantExampleOverwrite"
	testJobNameOverwrite2 = "levantExampleOverwrite2"
	testDCName            = "dc13"
	testEnvName           = "GROUP_NAME_ENV"
	testEnvValue          = "cache"
)

func TestTemplater_RenderTemplate(t *testing.T) {

	var job *nomad.Job
	var err error

	// Start with an empty passed var args map.
	fVars := make(map[string]interface{})

	// Test basic TF template render.
	job, err = RenderJob("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.tf"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if *job.TaskGroups[0].Tasks[0].Resources.CPU != 1313 {
		t.Fatalf("expected CPU resource %v but got %v", 1313, *job.TaskGroups[0].Tasks[0].Resources.CPU)
	}

	// Test basic YAML template render.
	job, err = RenderJob("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if *job.TaskGroups[0].Tasks[0].Resources.CPU != 1313 {
		t.Fatalf("expected CPU resource %v but got %v", 1313, *job.TaskGroups[0].Tasks[0].Resources.CPU)
	}

	// Test multiple var-files
	job, err = RenderJob("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.yaml", "test-fixtures/test-overwrite.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobNameOverwrite {
		t.Fatalf("expected %s but got %v", testJobNameOverwrite, *job.Name)
	}

	// Test multiple var-files of different types
	job, err = RenderJob("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.tf", "test-fixtures/test-overwrite.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobNameOverwrite {
		t.Fatalf("expected %s but got %v", testJobNameOverwrite, *job.Name)
	}

	// Test multiple var-files with var-args
	fVars["job_name"] = testJobNameOverwrite2
	job, err = RenderJob("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.tf", "test-fixtures/test-overwrite.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobNameOverwrite2 {
		t.Fatalf("expected %s but got %v", testJobNameOverwrite2, *job.Name)
	}

	// Test empty var-args and empty variable file render.
	job, err = RenderJob("test-fixtures/none_templated.nomad", []string{}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}

	// Test var-args only render.
	fVars = map[string]interface{}{"job_name": testJobName, "task_resource_cpu": "1313"}
	job, err = RenderJob("test-fixtures/single_templated.nomad", []string{}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if *job.TaskGroups[0].Tasks[0].Resources.CPU != 1313 {
		t.Fatalf("expected CPU resource %v but got %v", 1313, *job.TaskGroups[0].Tasks[0].Resources.CPU)
	}

	// Test var-args and variables file render.
	delete(fVars, "job_name")
	fVars["datacentre"] = testDCName
	os.Setenv(testEnvName, testEnvValue)
	job, err = RenderJob("test-fixtures/multi_templated.nomad", []string{"test-fixtures/test.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if job.Datacenters[0] != testDCName {
		t.Fatalf("expected %s but got %v", testDCName, job.Datacenters[0])
	}
	if *job.TaskGroups[0].Name != testEnvValue {
		t.Fatalf("expected %s but got %v", testEnvValue, *job.TaskGroups[0].Name)
	}
}
