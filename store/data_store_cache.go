package store

import (
	"encoding/gob"
	"os"
	"time"

	"github.com/AccumulatedFinance/aevm-bridge/config"
)

// Cache represents the snapshot of the data store.
type Cache struct {
	Blocks map[dskey]uint64
	Time   *time.Time
}

// WriteCache creates a cache (or snapshot) of the data store and stores it in the filesystem.
func (ds *DataStore) WriteCache() error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	now := time.Now()

	cacheData := &Cache{
		Blocks: ds.blocks,
		Time:   &now,
	}

	file, err := os.Create(ds.cacheFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(cacheData)
	if err != nil {
		return err
	}

	return nil
}

// ReadCache reads a cache (or snapshot) from the filesystem and applies it in-memory.
func (ds *DataStore) ReadCache(conf *config.Config) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	file, err := os.Open(ds.cacheFile)
	if err != nil {
		// Cache file not found or other error
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var cache Cache
	err = decoder.Decode(&cache)
	if err != nil {
		return err
	}

	if cache.Blocks == nil {
		cache.Blocks = make(map[dskey]uint64)
	}

	ds.blocks = cache.Blocks

	return nil
}
