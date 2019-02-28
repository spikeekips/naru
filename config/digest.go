package config

import (
	"time"

	"github.com/spikeekips/cvc"

	"github.com/spikeekips/naru/common"
)

type Digest struct {
	cvc.BaseGroup
	Init          bool          `flag-help:"digest from begining"`
	RemoteBlock   uint64        `flag-help:"digest from this block"`
	Watch         bool          `flag-help:"keep watching"`
	WatchInterval time.Duration `flag-help:"interval for watching"`
}

func NewDigest() *Digest {
	return &Digest{
		WatchInterval: common.DefaultDigestWatchInterval,
	}
}
