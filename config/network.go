package config

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	sebakcommon "boscoin.io/sebak/lib/common"
	"github.com/spikeekips/cvc"

	"github.com/spikeekips/naru/common"
)

type Network struct {
	cvc.BaseGroup
	Bind *sebakcommon.Endpoint `flag-help:"bind address"`
	TLS  *TLSConfig
	Log  *NetworkLogs
}

type TLSConfig struct {
	cvc.BaseGroup
	Cert string `flag-help:"tls cert file"`
	Key  string `flag-help:"tls key file"`
}

func NewNetwork() *Network {
	return &Network{
		Bind: common.DefaultBind,
		TLS:  &TLSConfig{},
		Log:  NewNetworkLogs(),
	}
}

func (n Network) ParseBind(i string) (*sebakcommon.Endpoint, error) {
	return sebakcommon.ParseEndpoint(i)
}

func (n TLSConfig) ParseCert(i string) (string, error) {
	if len(i) < 1 {
		return "", nil
	}

	path := filepath.Join(common.CurrentDirectory, i)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err
	}

	pb, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	if p, _ := pem.Decode(pb); p == nil {
		return "", fmt.Errorf("tls cert: invalid pem file")
	}

	return i, nil
}

func (n TLSConfig) ParseKey(i string) (string, error) {
	if len(i) < 1 {
		return "", nil
	}

	path := filepath.Join(common.CurrentDirectory, i)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err
	}

	pb, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	if p, _ := pem.Decode(pb); p == nil {
		return "", fmt.Errorf("tls key: invalid pem file")
	}

	return i, nil
}

type NetworkLogs struct {
	cvc.BaseGroup
	HTTP  *LogConfig
	Error *LogConfig
}

func NewNetworkLogs() *NetworkLogs {
	return &NetworkLogs{
		HTTP:  NewLog(),
		Error: NewLog(),
	}
}
