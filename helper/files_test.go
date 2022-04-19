package helper

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestHelper_GetDefaultTmplFile(t *testing.T) {
	d1 := []byte("Levant Test Job File\n")

	cases := []struct {
		TmplFiles []string
		Output    string
	}{
		{
			[]string{"example.nomad", "example1.nomad"},
			"",
		},
		{
			[]string{"example.nomad"},
			"example.nomad",
		},
	}

	for _, tc := range cases {
		for _, f := range tc.TmplFiles {

			// Use write file as tmpfile adds a prefix which doesn't work with the
			// GetDefaultTmplFile function.
			err := ioutil.WriteFile(f, d1, 0600)
			if err != nil {
				t.Fatal(err)
			}
		}

		actual := GetDefaultTmplFile()

		// Call explicit Remove as the function is dependant on the number of files
		// in the target directory.
		for _, f := range tc.TmplFiles {
			os.Remove(f)
		}

		if !reflect.DeepEqual(actual, tc.Output) {
			t.Fatalf("got: %#v, expected %#v", actual, tc.Output)
		}
	}

}

func TestHelper_GetDefaultVarFile(t *testing.T) {
	d1 := []byte("Levant Test Variable File\n")

	cases := []struct {
		VarFile string
	}{
		{"levant.yaml"},
		{"levant.yml"},
		{"levant.tf"},
		{""},
	}

	for _, tc := range cases {
		if tc.VarFile != "" {

			// Use write file as tmpfile adds a prefix which doesn't work with the
			// GetDefaultTmplFile function.
			err := ioutil.WriteFile(tc.VarFile, d1, 0600)
			if err != nil {
				t.Fatal(err)
			}
		}

		actual := GetDefaultVarFile()
		if !reflect.DeepEqual(actual, tc.VarFile) {
			t.Fatalf("got: %#v, expected %#v", actual, tc.VarFile)
		}

		os.Remove(tc.VarFile)
	}
}

func TestHelper_GetJobspecFromFile(t *testing.T) {
	cases := []struct {
		JobFile string
		JobName string
	}{
		{"test-fixtures/demojob.json", "demojob"},
	}

	for i, tc := range cases {
		job, err := GetJobspecFromFile(tc.JobFile)
		if err != nil {
			t.Fatalf("case %d failed: %v", i, err)
		}

		if !reflect.DeepEqual(*job.Name, tc.JobName) {
			t.Fatalf("got: %#v, expected %#v", *job.Name, tc.JobName)
		}

	}
}

func TestHelper_GetJobspecFromIOReader(t *testing.T) {
	cases := []struct {
		JobFile string
		JobName string
	}{
		{"test-fixtures/demojob.json", "demojob"},
	}

	for i, tc := range cases {
		src, err := ioutil.ReadFile(tc.JobFile)
		if err != nil {
			t.Fatalf("case %d failed: %v", i, err)
		}

		job, err := GetJobspecFromIOReader(bytes.NewBuffer(src))
		if err != nil {
			t.Fatalf("case %d failed: %v", i, err)
		}

		if !reflect.DeepEqual(*job.Name, tc.JobName) {
			t.Fatalf("got: %#v, expected %#v", *job.Name, tc.JobName)
		}

	}
}
