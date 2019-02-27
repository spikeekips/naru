package config

import "github.com/spikeekips/cvc"

type Digest struct {
	cvc.BaseGroup
	Init        bool   `flag-help:"digest from begining"`
	RemoteBlock uint64 `flag-help:"digest from this block"`
	Watch       bool   `flag-help:"keep watching"`
}
