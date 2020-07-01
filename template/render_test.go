package template

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"
)

const (
	testJobName           = `levantExample`
	testJobNameOverwrite2 = `levantExampleOverwrite2`
	testDCName            = `dc13`
	testEnvName           = `GROUP_NAME_ENV`
	testEnvValue          = `cache`
)

func TestTemplater_RenderTemplate(t *testing.T) {

	var job *bytes.Buffer
	var err error

	// Start with an empty passed var args map.
	fVars := make(map[string]string)

	// Test basic TF template render.
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.tf"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExample" {`, job)

	// Test basic YAML template render.
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExample" {`, job)

	// Test multiple var-files
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.yaml", "test-fixtures/test-overwrite.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExampleOverwrite" {`, job)

	// Test multiple var-files of different types
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.tf", "test-fixtures/test-overwrite.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExampleOverwrite" {`, job)

	// Test multiple var-files with var-args
	fVars["job_name"] = testJobNameOverwrite2
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", []string{"test-fixtures/test.tf", "test-fixtures/test-overwrite.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExampleOverwrite2" {`, job)

	// Test empty var-args and empty variable file render.
	job, err = RenderTemplate("test-fixtures/none_templated.nomad", []string{}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExample" {`, job)

	// Test var-args only render.
	delete(fVars, "job_name")
	fVars["job_name"] = testJobName
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", []string{}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExample" {`, job)

	// Test var-args and variables file render.
	delete(fVars, "job_name")
	fVars["datacentre"] = testDCName
	os.Setenv(testEnvName, testEnvValue)
	job, err = RenderTemplate("test-fixtures/multi_templated.nomad", []string{"test-fixtures/test.yaml"}, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	testInJob(t, `job "levantExample" {`, job)
	testInJob(t, `datacenters = \["dc13"\]`, job)
	testInJob(t, `group "cache" {`, job)
}

func testInJob(t *testing.T, pattern string, job *bytes.Buffer) {
	if !regexp.MustCompile(pattern).MatchString(job.String()) {
		fmt.Println(job.String())
		t.Fatalf("expected to find %s in job", pattern)
	}
}
