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
	idx := strings.Index(raw, "=")
	if idx == -1 {
		return fmt.Errorf("no '=' value in arg: %s", raw)
	}

	if *v == nil {
		*v = make(map[string]interface{})
	}

	// Split the variable key based on the nested delimiter to get a list of
	// nested keys.
	keys := strings.Split(raw[0:idx], ".")

	// If we only have a single key, then we are not dealing with a nested set
	// meaning we can update the variable mapping and exit.
	if len(keys) == 1 {
		(*v)[keys[0]] = raw[idx+1:]
		return nil
	}

	// Identify the index max of the list for easy use.
	nestedLen := len(keys) - 1

	// The end map is the only thing we concretely know which contains our
	// final key:value pair.
	endEntry := map[string]interface{}{keys[nestedLen]: raw[idx+1:]}

	// Track the root of the nested map structure so we can continue to iterate
	// the nested keys below.
	root := endEntry

	// Iterate the nested keys backwards. Set a new root map containing the
	// previous root as its value. Do not iterate backwards fully to the end,
	// instead save the first key for the entry into Flag.
	for i := nestedLen - 1; i > 0; i-- {
		root = map[string]interface{}{keys[i]: root}
	}
	(*v)[keys[0]] = root
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
