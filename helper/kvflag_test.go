// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
)

func TestHelper_Set(t *testing.T) {
	cases := []struct {
		Label  string
		Inputs []string
		Output map[string]interface{}
		Error  bool
	}{
		{
			"simple value",
			[]string{"key=value"},
			map[string]interface{}{"key": "value"},
			false,
		},
		{
			"nested replaces simple",
			[]string{"key=1", "key.nested=2"},
			nil,
			true,
		},
		{
			"simple replaces nested",
			[]string{"key.nested=2", "key=1"},
			map[string]interface{}{"key": "1"},
			false,
		},
		{
			"nested siblings",
			[]string{"nested.a=1", "nested.b=2"},
			map[string]interface{}{"nested": map[string]interface{}{"a": "1", "b": "2"}},
			false,
		},
		{
			"nested singleton",
			[]string{"nested.key=value"},
			map[string]interface{}{"nested": map[string]interface{}{"key": "value"}},
			false,
		},
		{
			"nested with parent",
			[]string{"root=a", "nested.key=value"},
			map[string]interface{}{"root": "a", "nested": map[string]interface{}{"key": "value"}},
			false,
		},
		{
			"empty value",
			[]string{"key="},
			map[string]interface{}{"key": ""},
			false,
		},
		{
			"value contains equal sign",
			[]string{"key=foo=bar"},
			map[string]interface{}{"key": "foo=bar"},
			false,
		},
		{
			"missing equal sign",
			[]string{"key"},
			nil,
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Label, func(t *testing.T) {
			f := new(Flag)
			mErr := multierror.Error{}
			for _, input := range tc.Inputs {
				err := f.Set(input)
				if err != nil {
					mErr.Errors = append(mErr.Errors, err)
				}
			}
			if tc.Error {
				require.Error(t, mErr.ErrorOrNil())
			} else {
				actual := map[string]interface{}(*f)
				require.True(t, reflect.DeepEqual(actual, tc.Output))
			}
		})
	}
}
