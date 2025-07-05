package store

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/AccumulatedFinance/aevm-bridge/config"
	"github.com/AccumulatedFinance/aevm-bridge/validation"
	"github.com/go-playground/validator/v10"
)

var Data *DataStore

type DataStore struct {
	mu        sync.RWMutex
	cacheFile string
	validate  *validator.Validate
	blocks    map[dskey]uint64
}

type dskey struct {
	ChainID int
	Address string
}

func NewDataStore(cacheFile string, conf *config.Config) *DataStore {

	ds := &DataStore{
		cacheFile: cacheFile,
		validate:  validation.GetInstance(),
		blocks:    make(map[dskey]uint64),
	}

	log.WithField("prefix", "store").Debug("reading cache file: ", cacheFile)
	if err := ds.ReadCache(conf); err != nil {
		log.Error(err)
	}

	return ds

}

func (st *DataStore) AddBlock(lastBlock uint64, chainId int, address string) {

	st.mu.Lock()
	defer st.mu.Unlock()

	st.blocks[dskey{chainId, strings.ToLower(address)}] = lastBlock
}

func (st *DataStore) GetBlock(chainId int, address string) (uint64, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	blockNumber, exists := st.blocks[dskey{chainId, strings.ToLower(address)}]
	if !exists {
		return 0, fmt.Errorf("block with address=%s and chainId=%d not found", address, chainId)
	}
	return blockNumber, nil
}
