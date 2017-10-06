package levant

import (
	"strings"
	"testing"

	nomad "github.com/hashicorp/nomad/api"
)

const (
	testJobName = "levantExample"
	testDCName  = "dc13"
)

func TestTemplater_RenderTemplate(t *testing.T) {

	var job *nomad.Job
	var err error

	// Start with an empty passed var args map.
	fVars := make(map[string]string)

	// Test basic TF template render.
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", "test-fixtures/test.tf", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}

	// Test basic YAML template render.
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", "test-fixtures/test.yaml", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}

	// Test empty var-args and empty variable file render.
	job, err = RenderTemplate("test-fixtures/none_templated.nomad", "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}

	// Test var-args only render.
	fVars["job_name"] = testJobName
	job, err = RenderTemplate("test-fixtures/single_templated.nomad", "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}

	// Test var-args and variables file render.
	delete(fVars, "job_name")
	fVars["datacentre"] = testDCName
	job, err = RenderTemplate("test-fixtures/multi_templated.nomad", "test-fixtures/test.yaml", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if job.Datacenters[0] != testDCName {
		t.Fatalf("expected %s but got %v", testDCName, job.Datacenters[0])
	}

	// Test var-args only render.
	fVars["job_name"] = testJobName
	job, err = RenderTemplate("test-fixtures/missing_var.nomad", "", &fVars)
	if err == nil {
		t.Fatal("expected err to not be nil")
	}
	if !strings.Contains(err.Error(), "binary_url") {
		t.Fatal("expected err to mention missing var (binary_url)")
	}
}
