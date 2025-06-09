// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package template

import (
	"os"
	"strings"
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

func findService(task *nomad.Task, portLabel string) (*nomad.Service, bool) {
	for _, service := range task.Services {
		if portLabel == service.PortLabel {
			return service, true
		}
	}
	return nil, false
}

// Test templates composed of other templates via the include function.
func TestTemplater_RenderTemplateInclude(t *testing.T) {
	compositionTasks := []struct {
		Name     string
		Image    string
		Memory   uint64
		Services map[string]int // name: port
	}{
		{"task1", "registry/task1:v1.1", 250, map[string]int{"http": 80, "https": 443}},
		{"task2", "registry/task2:v1.2", 300, map[string]int{"metrics": 8080}},
	}

	fVars := map[string]interface{}{
		"tasks": compositionTasks,
	}

	job, err := RenderJob("test-fixtures/composition_templated.nomad", []string{"test-fixtures/test.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	for i, expectedTask := range compositionTasks {
		actualTask := job.TaskGroups[0].Tasks[i]
		if actualTask.Name != expectedTask.Name {
			t.Fatalf("expected %s but got %v", expectedTask.Name, actualTask.Name)
		}
		actualTaskImage := actualTask.Config["image"].(string)
		if actualTaskImage != expectedTask.Image {
			t.Fatalf("expected %s but got %v", expectedTask.Image, actualTaskImage)
		}

		actualTaskPorts := actualTask.Config["port_map"].([]map[string]interface{})[0]
		for portName, expectedPort := range expectedTask.Services {
			if actualPort, ok := actualTaskPorts[portName]; !ok {
				t.Fatalf("expected %s in port_map of task %v", portName, expectedTask.Name)
			} else if actualPort.(int) != expectedPort {
				t.Fatalf("expected port_map[%s]=%v but got %v", portName, expectedPort, actualPort)
			}

			actualService, found := findService(actualTask, portName)
			if !found {
				t.Fatalf("expected %s in services of task %v", portName, expectedTask.Name)
			}
			expectedServiceName := "global-" + portName + "-check"
			if actualService.Name != expectedServiceName {
				t.Fatalf("expected service %s but got %v", expectedServiceName, actualService.Name)
			}
		}
	}

	// Test that cyclic includes are detected as an error.
	job, err = RenderJob("test-fixtures/recursive_include_template_1.nomad", nil, "", &fVars)
	if err == nil {
		t.Fatalf("expected error on cyclic includes")
	} else if !strings.Contains(err.Error(), "cyclic include detected") {
		t.Fatalf("expected error to contain 'cyclic include detected' but got %v", err)
	}
}
