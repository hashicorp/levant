// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scale

import (
	"testing"

	"github.com/hashicorp/levant/levant/structs"
	nomad "github.com/hashicorp/nomad/api"
)

func TestScale_updateTaskGroup(t *testing.T) {

	sOut := structs.ScalingDirectionOut
	sIn := structs.ScalingDirectionIn
	sCount := structs.ScalingDirectionTypeCount
	sPercent := structs.ScalingDirectionTypePercent

	cases := []struct {
		Config   *Config
		Group    *nomad.TaskGroup
		EndCount int
	}{
		{
			buildScalingConfig(sOut, sCount, 100),
			buildTaskGroup(1000),
			1100,
		},
		{
			buildScalingConfig(sOut, sPercent, 25),
			buildTaskGroup(100),
			125,
		},
		{
			buildScalingConfig(sIn, sCount, 900),
			buildTaskGroup(901),
			1,
		},
		{
			buildScalingConfig(sIn, sPercent, 90),
			buildTaskGroup(100),
			10,
		},
	}

	for _, tc := range cases {
		updateTaskGroup(tc.Config, tc.Group)

		if tc.EndCount != *tc.Group.Count {
			t.Fatalf("got: %#v, expected %#v", *tc.Group.Count, tc.EndCount)
		}
	}
}

func TestScale_calculateCountBasedOnPercent(t *testing.T) {

	cases := []struct {
		Count   int
		Percent int
		Output  int
	}{
		{
			100,
			50,
			50,
		},
		{
			3,
			33,
			1,
		},
		{
			3,
			10,
			0,
		},
	}

	for _, tc := range cases {
		output := calculateCountBasedOnPercent(tc.Count, tc.Percent)

		if output != tc.Output {
			t.Fatalf("got: %#v, expected %#v", output, tc.Output)
		}
	}
}

func buildScalingConfig(direction, dType string, number int) *Config {

	c := &Config{
		Scale: &structs.ScaleConfig{
			Direction:     direction,
			DirectionType: dType,
		},
	}

	switch dType {
	case structs.ScalingDirectionTypeCount:
		c.Scale.Count = number
	case structs.ScalingDirectionTypePercent:
		c.Scale.Percent = number
	}

	return c
}

func buildTaskGroup(count int) *nomad.TaskGroup {

	n := "LevantTest"
	c := count

	t := &nomad.TaskGroup{
		Name:  &n,
		Count: &c,
	}

	return t
}
