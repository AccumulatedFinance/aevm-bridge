package main

import (
	"flag"
	"os/user"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/AccumulatedFinance/aevm-bridge/config"
	"github.com/AccumulatedFinance/aevm-bridge/evm"
	"github.com/AccumulatedFinance/aevm-bridge/store"
)

const CACHE_FILE = "cache.gob"

func main() {

	usr, err := user.Current()
	if err != nil {
		log.WithField("prefix", "main").Error(err)
	}

	dir := usr.HomeDir + "/.aevm"
	flag.StringVar(&dir, "dir", dir, "dir path")

	flag.Parse()

	start(dir)

}

func start(dir string) {

	var err error
	var conf *config.Config

	// load config
	if conf, err = config.NewConfig(dir); err != nil {
		log.WithField("prefix", "main").Fatal(err)
	}

	// init stores
	log.WithField("prefix", "main").Debug("initializing data store")
	store.Data = store.NewDataStore(filepath.Join(dir, CACHE_FILE), conf)
	store.EVM = store.NewEVMStore()

	// init evm network
	for _, v := range conf.EVMNetworks {
		client, err := evm.NewEVMClient(v.ChainID, v.Endpoint)
		if err != nil {
			log.WithField("prefix", "main").Fatal(err)
		}
		store.EVM.AddClient(client)
	}

	die := make(chan bool)

	// init bridges
	bridges := &conf.Bridges
	for _, b := range *bridges {
		go getBridge(b, die)
	}

	// persistent cache of the data store every minute
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			err := store.Data.WriteCache()
			if err != nil {
				log.WithField("prefix", "main").Error("failed to write data store cache:", err)
			}
		}
	}()

	select {}

}

// getBridge parses presale deposit events
func getBridge(p config.Bridge, die chan bool) {

	log.Info("Parsing bridge on chain ", p.ChainID)

	var firstBlock uint64
	var lastBlock uint64

	bridgeBlock, err := store.Data.GetBlock(p.ChainID, p.Address)
	if err != nil {
		firstBlock = p.BlockNumber
	} else {
		firstBlock = bridgeBlock
		// if received higher blocknumber from config, ignore cache
		if p.BlockNumber > bridgeBlock {
			firstBlock = p.BlockNumber
		}
	}

	timeout := int64(5)
	//	val := validation.GetInstance()

	client, err := store.EVM.GetClientByChainId(p.ChainID)
	if err != nil {
		log.WithField("prefix", "main").Error(err)
		return
	}

	for {

		select {
		default:

			time.Sleep(time.Duration(timeout) * time.Second)

			// lastBlock always = first + LIMIT
			lastBlock = firstBlock + evm.EVM_EVENTS_LIMIT

			// check current evm block, if last block > current, use current as last instead
			currentBlock, err := client.GetCurrentBlockNumber()
			if err != nil {
				log.WithField("prefix", "main").Error("Error fetching current block number: ", err)
				timeout = 30
				continue
			}

			// Check if lastBlock is greater than currentBlock, if so, set lastBlock to currentBlock
			if lastBlock > currentBlock {
				lastBlock = currentBlock
			}

			log.Info("Parsing events on chain=", p.ChainID, " from ", firstBlock, " to ", lastBlock)

			evmEvents, err := client.GetDepositEvents(p.Address, firstBlock, &lastBlock)
			if err != nil {
				log.WithField("prefix", "main").Error(err)
				timeout = 30
				continue
			}

			log.Debug("Found ", len(evmEvents), " deposit events on chain ", p.ChainID)

			if len(evmEvents) > 0 {
				for _, event := range evmEvents {
					log.Info(event.Amount)
				}
			}

			store.Data.AddBlock(lastBlock, p.ChainID, p.Address)
			firstBlock = lastBlock

			timeout = 30

			// if we parse history events from old blocks, accelerate parsing
			if lastBlock != currentBlock {
				timeout = 3
			}

		case <-die:
			return
		}

	}

}
