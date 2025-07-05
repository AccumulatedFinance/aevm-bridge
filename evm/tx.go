package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// PrepareTx fills gas params for legacyTx depending on the txType of the client
func (e *EVMClient) PrepareTx(legacyTx *types.LegacyTx) *types.Transaction {

	tx := &types.Transaction{}

	if e.TxType == 0 {
		legacyTx.GasPrice = e.GasFeeCap
		legacyTx.Gas = e.GasLimit
		tx = types.NewTx(legacyTx)
	}

	if e.TxType == 1 {
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(int64(e.ChainID)),
			GasFeeCap: e.GasFeeCap,
			GasTipCap: e.GasTipCap,
			Gas:       e.GasLimit,
			To:        legacyTx.To,
			Value:     legacyTx.Value,
			Data:      legacyTx.Data,
			Nonce:     legacyTx.Nonce,
		})
	}

	return tx

}

// SubmitTx adds signature and submits tx
func (e *EVMClient) SubmitTx(tx *types.Transaction, signature []byte) (common.Hash, error) {

	signedTx, err := tx.WithSignature(types.LatestSignerForChainID(big.NewInt(int64(e.ChainID))), signature)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to apply signature: %w", err)
	}

	if err = e.Client.SendTransaction(context.Background(), signedTx); err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), nil

}
