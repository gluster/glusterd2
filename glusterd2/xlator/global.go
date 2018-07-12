package xlator

import (
	"fmt"
	"runtime/debug"

	"github.com/gluster/glusterd2/glusterd2/xlator/options"

	log "github.com/sirupsen/logrus"
)

var (
	// xlMap is a map of all available xlators, indexed by xlator-id
	xlMap map[string]*Xlator
	// optMap is a map of all available volume options indexed by
	// <xlator-id>.<option-key> for all keys of a volume option
	optMap map[string]*options.Option
)

// Load load all available xlators and intializes the xlators and options maps
func Load() (err error) {

	defer func() {
		if r := recover(); r != nil {
			log.Info(string(debug.Stack()))
			err = fmt.Errorf("recover()ed at xlator.Load(): %s", r)
			log.Error("Your version of glusterfs is incompatible. ",
				"Please install latest glusterfs from source (branch: master)")
		}
	}()

	xls, err := loadAllXlators()
	if err != nil {
		return
	}
	xlMap = xls

	injectTransportOptions()
	loadOptions()
	return
}

// Xlators returns the xlator map
func Xlators() map[string]*Xlator {
	return xlMap
}

// loadOptions loads all available options into the options.Options map,
// indexed as <xlator-id>.<option-key> for all available option keys
func loadOptions() {
	optMap = make(map[string]*options.Option)
	for _, xl := range xlMap {
		for _, opt := range xl.Options {
			for _, k := range opt.Key {
				k := xl.ID + "." + k
				optMap[k] = opt
			}
		}
	}
}

// injectTransportOptions injects options present in transport layer (socket.so
// and rdma.so) into list of options loaded from protocol layer (server.so and
// client.so)
func injectTransportOptions() {

	var transportNames = [...]string{"socket", "rdma"}
	transports := make([]*Xlator, 0, 2)
	for _, name := range transportNames {
		if xl, ok := xlMap[name]; ok {
			transports = append(transports, xl)
		}
	}

	if len(transports) == 0 {
		panic("socket.so or rdma.so not found. Please install glusterfs-server package")
	}

	for _, transport := range transports {
		for _, option := range transport.Options {
			// TODO:
			// remove this once proper settable flags are set for
			// these transport options in glusterfs source
			option.Flags = option.Flags | options.OptionFlagSettable
			if xl, ok := xlMap["server"]; ok {
				xl.Options = append(xl.Options, option)
			}
			if xl, ok := xlMap["client"]; ok {
				option.Flags = option.Flags | options.OptionFlagClientOpt
				xl.Options = append(xl.Options, option)
			}
		}
	}
}
