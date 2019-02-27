package config

import "github.com/spikeekips/cvc"

type System struct {
	cvc.BaseGroup
	Profile bool `flag-help:"enable profiling"`
}

func NewSystem() *System {
	return &System{}
}
