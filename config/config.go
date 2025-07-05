package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/configor"
	"github.com/mcuadros/go-defaults"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	PrivateKey  string      `yaml:"privateKey" json:"privateKey" form:"privateKey" query:"privateKey"`
	EVMNetworks EVMNetworks `yaml:"evmNetworks" json:"evmNetworks" form:"evmNetworks" query:"evmNetworks"`
	Bridges     []Bridge    `yaml:"bridges" json:"bridges" form:"bridges" query:"bridges"`
}

type EVMNetworks []EVMNetwork

type EVMNetwork struct {
	ChainID   int             `required:"true" yaml:"chainID" json:"chainID" form:"chainID" query:"chainID"`
	Endpoint  string          `required:"true" yaml:"endpoint" json:"endpoint" form:"endpoint" query:"endpoint"`
	Coin      *EVMNetworkCoin `yaml:"coin" json:"coin" form:"coin" query:"coin"`
	TxType    int             `default:"1" yaml:"txType" json:"txType" form:"txType" query:"txType"`               // default 0, for Ethereum will be 1
	GasLimit  int64           `default:"5000000" yaml:"gasLimit" json:"gasLimit" form:"gasLimit" query:"gasLimit"` // default 50000000 for all chains, rewrite if needed
	GasFeeCap float64         `yaml:"gasFeeCap" json:"gasFeeCap" form:"gasFeeCap" query:"gasFeeCap"`               // parse from blockchain by default
	GasTipCap float64         `default:"0.1" yaml:"gasTipCap" json:"gasTipCap" form:"gasTipCap" query:"gasTipCap"` // default 0.1 for Ethereum, not used on most other chains

}

type EVMNetworkCoin struct {
	Symbol   string `required:"true" yaml:"symbol" json:"symbol" form:"symbol" query:"symbol"`
	Decimals int    `required:"true" yaml:"decimals" json:"decimals" form:"decimals" query:"decimals"`
}

type Bridge struct {
	ChainID     int    `required:"true" yaml:"chainID" json:"chainID" form:"chainID" query:"chainID"`
	Address     string `required:"true" yaml:"address" json:"address" form:"address" query:"address"`
	RebaseToken string `required:"true" yaml:"rebaseToken" json:"rebaseToken" form:"rebaseToken" query:"rebaseToken"`
	BlockNumber uint64 `yaml:"blockNumber" json:"blockNumber" form:"blockNumber" query:"blockNumber"`
}

// NewConfig creates config from configFile
func NewConfig(configPath string) (*Config, error) {

	config := new(Config)
	defaults.SetDefaults(config)

	var configBytes []byte

	err := filepath.WalkDir(configPath, func(filePath string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Ignore folders starting with "_"
		if info.IsDir() && strings.HasPrefix(info.Name(), "_") {
			return filepath.SkipDir
		}
		// Check if the file or folder has a valid ".yaml" extension
		if strings.HasSuffix(info.Name(), ".yaml") && !strings.HasPrefix(info.Name(), "_") {
			log.Info(filePath)
			cBytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}
			configBytes = append(configBytes, cBytes...)
			// Add an empty line after each YAML file
			configBytes = append(configBytes, '\n')
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}

	if err := configor.Load(config); err != nil {
		return nil, err
	}

	return config, nil
}

// getCoinByChainID finds network coin by chainID
func (networks *EVMNetworks) GetCoinByChainID(chainID int) (*EVMNetworkCoin, error) {
	for _, network := range *networks {
		if network.ChainID == chainID {
			if network.Coin != nil {
				return network.Coin, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find coin for chainID %d", chainID)
}
