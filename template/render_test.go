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

func TestTemplater_DeepMergeVariables(t *testing.T) {
	// Test deep merge functionality with nested variables
	fVars := make(map[string]interface{})
	
	job, err := RenderJob("test-fixtures/nested_templated.nomad", []string{"test-fixtures/test-nested-1.yaml", "test-fixtures/test-nested-2.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	
	// Check that the job config was properly rendered (this validates the deep merge worked)
	config := job.TaskGroups[0].Tasks[0].Config
	args, ok := config["args"].([]interface{})
	if !ok {
		t.Fatal("expected args to be a slice")
	}
	
	// Check that we have the expected number of args
	if len(args) != 7 {
		t.Fatalf("expected 7 args but got %d", len(args))
	}
	
	// Check that database config from both files is present
	found := make(map[string]bool)
	expected := []string{"DB Host: localhost", "DB Port: 5432", "DB User: admin", "DB Pass: secret", "Cache Enabled: true", "Cache TTL: 300", "Log Level: debug"}
	
	for _, arg := range args {
		argStr := arg.(string)
		for _, exp := range expected {
			if argStr == exp {
				found[exp] = true
			}
		}
	}
	
	for _, exp := range expected {
		if !found[exp] {
			t.Errorf("expected argument '%s' not found in job config. Got args: %v", exp, args)
		}
	}
}

func TestDeepMerge(t *testing.T) {
	// Test the deepMerge function directly
	dst := map[string]interface{}{
		"config": map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
			"cache": map[string]interface{}{
				"enabled": true,
				"ttl":     300,
			},
		},
		"job_name": "levantExample",
	}
	
	src := map[string]interface{}{
		"config": map[string]interface{}{
			"database": map[string]interface{}{
				"username": "admin",
				"password": "secret",
			},
			"logging": map[string]interface{}{
				"level": "debug",
			},
		},
	}
	
	deepMerge(dst, src)
	
	// Check that the merge worked correctly
	config, ok := dst["config"].(map[string]interface{})
	if !ok {
		t.Fatal("config should be a map")
	}
	
	database, ok := config["database"].(map[string]interface{})
	if !ok {
		t.Fatal("database should be a map")
	}
	
	// Check all database fields are present
	if database["host"] != "localhost" {
		t.Errorf("expected host localhost, got %v", database["host"])
	}
	if database["port"] != 5432 {
		t.Errorf("expected port 5432, got %v", database["port"])
	}
	if database["username"] != "admin" {
		t.Errorf("expected username admin, got %v", database["username"])
	}
	if database["password"] != "secret" {
		t.Errorf("expected password secret, got %v", database["password"])
	}
	
	// Check cache section is preserved
	cache, ok := config["cache"].(map[string]interface{})
	if !ok {
		t.Fatal("cache should be a map")
	}
	if cache["enabled"] != true {
		t.Errorf("expected cache enabled true, got %v", cache["enabled"])
	}
	if cache["ttl"] != 300 {
		t.Errorf("expected cache ttl 300, got %v", cache["ttl"])
	}
	
	// Check logging section is added
	logging, ok := config["logging"].(map[string]interface{})
	if !ok {
		t.Fatal("logging should be a map")
	}
	if logging["level"] != "debug" {
		t.Errorf("expected logging level debug, got %v", logging["level"])
	}
}

func TestTemplater_DeepMergeJSON(t *testing.T) {
	// Test deep merge functionality with JSON files
	fVars := make(map[string]interface{})
	
	job, err := RenderJob("test-fixtures/nested_templated.nomad", []string{"test-fixtures/test-nested-1.json", "test-fixtures/test-nested-2.json"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	
	// Verify the nested variables were properly merged for JSON
	config := job.TaskGroups[0].Tasks[0].Config
	args, ok := config["args"].([]interface{})
	if !ok {
		t.Fatal("expected args to be a slice")
	}
	
	// Check that database config from both files is present
	found := make(map[string]bool)
	expected := []string{"DB Host: localhost", "DB Port: 5432", "DB User: admin", "DB Pass: secret", "Cache Enabled: true", "Cache TTL: 300", "Log Level: debug"}
	
	for _, arg := range args {
		argStr := arg.(string)
		for _, exp := range expected {
			if argStr == exp {
				found[exp] = true
			}
		}
	}
	
	for _, exp := range expected {
		if !found[exp] {
			t.Errorf("expected argument '%s' not found in job config. Got args: %v", exp, args)
		}
	}
}

func TestTemplater_DeepMergeTerraform(t *testing.T) {
	// Test deep merge functionality with Terraform files
	fVars := make(map[string]interface{})
	
	job, err := RenderJob("test-fixtures/nested_templated.nomad", []string{"test-fixtures/test-nested-1.tf", "test-fixtures/test-nested-2.tf"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	
	// Verify the nested variables were properly merged for Terraform
	config := job.TaskGroups[0].Tasks[0].Config
	args, ok := config["args"].([]interface{})
	if !ok {
		t.Fatal("expected args to be a slice")
	}
	
	// Check that database config from both files is present
	found := make(map[string]bool)
	expected := []string{"DB Host: localhost", "DB Port: 5432", "DB User: admin", "DB Pass: secret", "Cache Enabled: true", "Cache TTL: 300", "Log Level: debug"}
	
	for _, arg := range args {
		argStr := arg.(string)
		for _, exp := range expected {
			if argStr == exp {
				found[exp] = true
			}
		}
	}
	
	for _, exp := range expected {
		if !found[exp] {
			t.Errorf("expected argument '%s' not found in job config. Got args: %v", exp, args)
		}
	}
}
