package structs

// These consts represent configuration indentifers to use when performing
// either a scale-out or scale-in operation.
const (
	ScalingDirectionOut         = "Out"
	ScalingDirectionIn          = "In"
	ScalingDirectionTypeCount   = "Count"
	ScalingDirectionTypePercent = "Percent"
)

// ScalingConfig is an internal config struct used to track configuration
// details when performing a scale-out or scale-in operation.
type ScalingConfig struct {
	// Addr is the Nomad API address to use for all calls and must include both
	// protocol and port.
	Addr string

	// Count is the count by which the operator has asked to scale the Nomad job
	// and optional taskgroup by.
	Count int

	// Direction is the direction in which the scaling will take place and is
	// populated by consts.
	Direction string

	// DirectionType is an identifier on whether the operator has specified to
	// scale using a count increase or percentage.
	DirectionType string

	// JobID is the Nomad job which will be interacted with for scaling.
	JobID string

	// Percent is the percentage by which the operator has asked to scale the
	// Nomad job and optional taskgroup by.
	Percent int

	// TaskGroup is the Nomad job taskgroup which has been selected for scaling.
	TaskGroup string
}
