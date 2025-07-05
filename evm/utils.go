package evm

import "strings"

// isExecutionReverted checks if the error is a contract revert
func isExecutionReverted(err error) bool {
	return strings.Contains(err.Error(), "execution reverted")
}
