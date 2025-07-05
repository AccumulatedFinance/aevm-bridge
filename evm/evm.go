package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/AccumulatedFinance/aevm-bridge/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const GWEI = 1e9

const EVM_CLIENT_DELAY = 100 * time.Millisecond
const EVM_EVENTS_LIMIT = 29
const ETH_ZERO_ADDRESS = "0x0000000000000000000000000000000000000000"
const ETH_ETHER_ADDRESS = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
const ETH_ETHER_ADDRESS_MIXED = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"

type EVMClient struct {
	ChainID    int               `json:"chainId" validate:"number,gt=0"`
	TxType     int               // 0 or 1
	Client     *ethclient.Client `json:"client"`
	PrivateKey *ecdsa.PrivateKey
	PublicKey  common.Address
	GasFeeCap  *big.Int
	GasTipCap  *big.Int
	GasLimit   uint64
}

// NewEVMClient constructs the EVM client
func NewEVMClient(conf config.EVMNetwork) (*EVMClient, error) {

	c := &EVMClient{}
	c.ChainID = conf.ChainID

	client, err := ethclient.Dial(conf.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("can not connect to node: %s", conf.Endpoint)
	}

	c.Client = client

	// load gas params from config
	c.GasLimit = uint64(conf.GasLimit)

	gasFeeCap := big.NewFloat(conf.GasFeeCap)
	gasFeeCap.Mul(gasFeeCap, big.NewFloat(GWEI))
	c.GasFeeCap, _ = gasFeeCap.Int(nil)

	// if no GasFeeCap set, parse from blockchain
	if c.GasFeeCap.Cmp(big.NewInt(0)) == 0 {
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, err
		}
		c.GasFeeCap = gasPrice
	}

	// tx type 1
	if c.TxType == 1 {
		gasTipCap := big.NewFloat(conf.GasTipCap)
		gasTipCap.Mul(gasTipCap, big.NewFloat(GWEI))
		c.GasTipCap, _ = gasTipCap.Int(nil)
		// if no GasTipCap set, parse from blockchain
		if c.GasTipCap.Cmp(big.NewInt(0)) == 0 {
			gasTip, err := client.SuggestGasTipCap(context.Background())
			if err != nil {
				return nil, err
			}
			c.GasTipCap = gasTip
		}
	}

	return c, nil

}

// GetCurrentBlockNumber returns the current block number of the connected Ethereum chain.
func (e *EVMClient) GetCurrentBlockNumber() (uint64, error) {
	// Fetch the latest block number using the Ethereum client
	blockNumber, err := e.Client.BlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("can not fetch block number: %s", err)
	}
	return blockNumber, nil
}

// ImportPrivateKey imports private key and generates corresponding public key
func (e *EVMClient) ImportPrivateKey(pk string) (*EVMClient, error) {

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, err
	}

	e.PrivateKey = privateKey

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	e.PublicKey = crypto.PubkeyToAddress(*publicKeyECDSA)

	return e, nil

}
