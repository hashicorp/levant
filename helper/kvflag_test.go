package helper

import (
	"reflect"
	"testing"
)

func TestHelper_Set(t *testing.T) {
	cases := []struct {
		Input  string
		Output map[string]interface{}
		Error  bool
	}{
		{
			"key=value",
			map[string]interface{}{"key": "value"},
			false,
		},
		{
			"nested.key=value",
			map[string]interface{}{"nested": map[string]interface{}{"key": "value"}},
			false,
		},
		{
			"key=",
			map[string]interface{}{"key": ""},
			false,
		},
		{
			"key=foo=bar",
			map[string]interface{}{"key": "foo=bar"},
			false,
		},
		{
			"key",
			nil,
			true,
		},
	}

	for _, tc := range cases {
		f := new(Flag)
		err := f.Set(tc.Input)
		if (err != nil) != tc.Error {
			t.Fatalf("bad error. Input: %#v", tc.Input)
		}

		actual := map[string]interface{}(*f)
		if !reflect.DeepEqual(actual, tc.Output) {
			t.Fatalf("bad: %#v", actual)
		}
	}
}
