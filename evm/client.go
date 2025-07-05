package evm

import (
	"context"
	"math/big"

	"github.com/AccumulatedFinance/aevm-bridge/binding"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type BridgeEvent struct {
	// event log raw
	Timestamp   uint64 `json:"timestamp" validate:"number,gt=0"`
	BlockNumber uint64 `json:"blockNumber" validate:"number,gt=0"`
	TxHash      string `json:"txHash" validate:"required,eth_bytes32"`
	// event log data
	Receiver string   `json:"receiver" validate:"required,eth_addr"`
	Amount   *big.Int `json:"amount" validate:"required,number,gte=0"`
}

// GetDepositEvents retrieves Deposit events from start to end blocks
func (e *EVMClient) GetDepositEvents(address string, start uint64, end *uint64) ([]*BridgeEvent, error) {

	var events []*BridgeEvent

	bridgeAddress := common.HexToAddress(address)
	instance, err := binding.NewBridge(bridgeAddress, e.Client)
	if err != nil {
		return nil, err
	}

	opts := &bind.FilterOpts{
		Start:   start,
		End:     end,
		Context: context.Background(),
	}

	iter, err := instance.FilterDeposit(opts, nil) // nil values match all receivers/tokens
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.Next() {
		blockNumber := iter.Event.Raw.BlockNumber
		header, err := e.Client.HeaderByNumber(context.Background(), big.NewInt(int64(blockNumber)))
		if err != nil {
			return nil, err
		}
		events = append(events, &BridgeEvent{
			Timestamp:   header.Time,
			BlockNumber: iter.Event.Raw.BlockNumber,
			TxHash:      iter.Event.Raw.TxHash.Hex(),
			Receiver:    iter.Event.Receiver.Hex(),
			Amount:      iter.Event.Amount,
		})
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}

	return events, nil
}
