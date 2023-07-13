// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"fmt"
	"strings"
)

// Flag is a flag.Value implementation for parsing user variables
// from the command-line in the format of '-var key=value'.
type Flag map[string]interface{}

func (v *Flag) String() string {
	return ""
}

// Set takes a flag variable argument and pulls the correct key and value to
// create or add to a map.
func (v *Flag) Set(raw string) error {
	split := strings.SplitN(raw, "=", 2)
	if len(split) != 2 {
		return fmt.Errorf("no '=' value in arg: %s", raw)
	}
	keyRaw, value := split[0], split[1]

	if *v == nil {
		*v = make(map[string]interface{})
	}

	// Split the variable key based on the nested delimiter to get a list of
	// nested keys.
	keys := strings.Split(keyRaw, ".")

	lastKeyIdx := len(keys) - 1
	// Find the nested map where this value belongs
	// create missing maps as we go
	target := *v
	for i := 0; i < lastKeyIdx; i++ {
		raw, ok := target[keys[i]]
		if !ok {
			raw = make(map[string]interface{})
			target[keys[i]] = raw
		}
		var newTarget Flag
		if newTarget, ok = raw.(map[string]interface{}); !ok {
			return fmt.Errorf("simple value already exists at key %q", strings.Join(keys[:i+1], "."))
		}
		target = newTarget
	}
	target[keys[lastKeyIdx]] = value
	return nil
}

// FlagStringSlice is a flag.Value implementation for parsing targets from the
// command line, e.g. -var-file=aa -var-file=bb
type FlagStringSlice []string

func (v *FlagStringSlice) String() string {
	return ""
}

// Set is used to append a variable file flag argument to a list of file flag
// args.
func (v *FlagStringSlice) Set(raw string) error {
	*v = append(*v, raw)
	return nil
}
