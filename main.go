package main

import (
	"flag"
	"os/user"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/AccumulatedFinance/aevm-bridge/config"
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
	var lvl log.Level

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
		// go getBridge(b, conf.EVMNetworks, die)
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

}
