package evm

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

func (e *EVMClient) GenerateAndSignERC20Mint(contractAddr, recipient common.Address, amount *big.Int) (*types.LegacyTx, []byte, error) {
	// Parse ERC20 ABI with mint(address,uint256)
	erc20ABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"}]`))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Encode call data
	data, err := erc20ABI.Pack("mint", recipient, amount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to pack mint call: %w", err)
	}

	// Get current nonce
	nonce, err := e.Client.PendingNonceAt(context.Background(), e.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Create legacy tx
	legacyTx := &types.LegacyTx{
		Nonce:    nonce,
		To:       &contractAddr,
		Value:    big.NewInt(0),
		Gas:      e.GasLimit,
		GasPrice: e.GasFeeCap,
		Data:     data,
	}

	// Wrap it into a full tx so we can hash & sign
	tx := types.NewTx(legacyTx)

	// Get signer for the current chain ID
	signer := types.LatestSignerForChainID(big.NewInt(int64(e.ChainID)))

	// Hash the tx
	txHash := signer.Hash(tx).Bytes()

	// Sign it
	signature, err := crypto.Sign(txHash, e.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	return legacyTx, signature, nil
}
