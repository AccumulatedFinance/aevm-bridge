package validation

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/go-playground/validator/v10"
)

var lock = &sync.Mutex{}

type Single struct {
	Validator *validator.Validate
}

var singleInstance *Single

// GetInstance is a singleton for validator (it creates new or re-uses earlier created instance)
func GetInstance() *validator.Validate {

	if singleInstance == nil {

		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			log.Debug("creating single instance")
			instance := validator.New()
			singleInstance = &Single{instance}
		} else {
			log.Debug("single instance already created")
		}

	} else {

		log.Debug("single instance already created")

	}
	return singleInstance.Validator
}
