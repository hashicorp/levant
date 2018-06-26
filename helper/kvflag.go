package helper

import (
	"fmt"
	"strings"
)

// Flag is a flag.Value implementation for parsing user variables
// from the command-line in the format of '-var key=value'.
type Flag map[string]string

func (v *Flag) String() string {
	return ""
}

// Set takes...
func (v *Flag) Set(raw string) error {
	idx := strings.Index(raw, "=")
	if idx == -1 {
		return fmt.Errorf("no '=' value in arg: %s", raw)
	}

	if *v == nil {
		*v = make(map[string]string)
	}

	key, value := raw[0:idx], raw[idx+1:]
	(*v)[key] = value
	return nil
}

// FlagStringSlice is a flag.Value implementation for parsing targets from the
// command line, e.g. -var-file=aa -var-file=bb
type FlagStringSlice []string

func (v *FlagStringSlice) String() string {
	return ""
}
func (v *FlagStringSlice) Set(raw string) error {
	*v = append(*v, raw)
	return nil
}
