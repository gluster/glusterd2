package transaction

// The StepFunc registry registers StepFunc's to be used by transaction framework

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

var sfRegistry = struct {
	sync.RWMutex
	sfMap map[string]StepFunc
}{}

func registerStepFunc(s StepFunc, name string) {
	if sfRegistry.sfMap == nil {
		sfRegistry.sfMap = make(map[string]StepFunc)
	}

	if _, ok := sfRegistry.sfMap[name]; ok {
		log.WithField("stepname", name).Warning("step with provided name exists in registry and will be overwritten")
	}

	sfRegistry.sfMap[name] = s
}

//RegisterStepFunc registers the given StepFunc in the registry
func RegisterStepFunc(s StepFunc, name string) {
	sfRegistry.Lock()
	defer sfRegistry.Unlock()

	registerStepFunc(s, name)
}

//getStepFunc returns named step if found.
func getStepFunc(name string) (StepFunc, bool) {
	sfRegistry.RLock()
	defer sfRegistry.RUnlock()

	s, ok := sfRegistry.sfMap[name]
	return s, ok
}
