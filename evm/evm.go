package evm

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

const EVM_CLIENT_DELAY = 100 * time.Millisecond
const EVM_EVENTS_LIMIT = 29
const ETH_ZERO_ADDRESS = "0x0000000000000000000000000000000000000000"
const ETH_ETHER_ADDRESS = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
const ETH_ETHER_ADDRESS_MIXED = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"

type EVMClient struct {
	ChainID int               `json:"chainId" validate:"number,gt=0"`
	Client  *ethclient.Client `json:"client"`
}

// NewEVMClient constructs the EVM client
func NewEVMClient(chainID int, endpoint string) (*EVMClient, error) {

	c := &EVMClient{}
	c.ChainID = chainID

	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("can not connect to node: %s", endpoint)
	}

	c.Client = client

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
