package config

import (
	sebakcommon "boscoin.io/sebak/lib/common"
	"github.com/spikeekips/cvc"

	"github.com/spikeekips/naru/common"
)

type SEBAK struct {
	cvc.BaseGroup
	Endpoint *sebakcommon.Endpoint `flag:"node" flag-help:"sebak endpoint"`
	JSONRpc  *sebakcommon.Endpoint `flag-help:"sebak jsonrpc endpoint"`
}

func NewSEBAK() *SEBAK {
	return &SEBAK{
		Endpoint: common.DefaultSEBAKEndpoint,
		JSONRpc:  common.DefaultSEBAKJSONRpc,
	}
}

func (s SEBAK) ParseEndpoint(i string) (*sebakcommon.Endpoint, error) {
	return sebakcommon.ParseEndpoint(i)
}

func (s SEBAK) ParseJSONRpc(i string) (*sebakcommon.Endpoint, error) {
	return sebakcommon.ParseEndpoint(i)
}
