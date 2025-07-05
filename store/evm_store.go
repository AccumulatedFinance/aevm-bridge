package store

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/AccumulatedFinance/aevm-bridge/evm"
	"github.com/AccumulatedFinance/aevm-bridge/validation"
	"github.com/go-playground/validator/v10"
)

var EVM *EVMStore

type EVMStore struct {
	mu       sync.RWMutex
	validate *validator.Validate
	clients  map[int]*evm.EVMClient
}

func NewEVMStore() *EVMStore {
	return &EVMStore{
		validate: validation.GetInstance(),
		clients:  make(map[int]*evm.EVMClient),
	}
}

func (st *EVMStore) AddClient(client *evm.EVMClient) {

	if err := st.validate.Struct(client); err != nil {
		log.Error(err)
	}

	st.mu.Lock()
	defer st.mu.Unlock()
	st.clients[client.ChainID] = client
}

func (st *EVMStore) GetClientByChainId(chainID int) (*evm.EVMClient, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	client, exists := st.clients[chainID]
	if !exists {
		return nil, fmt.Errorf("client with chainID %d not found", chainID)
	}

	return client, nil
}

func (st *EVMStore) GeAllClients() map[int]*evm.EVMClient {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.clients
}
