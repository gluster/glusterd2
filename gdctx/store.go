package gdctx

import (
	"github.com/gluster/glusterd2/store"

	log "github.com/Sirupsen/logrus"
)

// If someone needs to use the GD2 store, all they need to do is just import context and use context.Store
var (
	Store    *store.GDStore
	prefixes []string
)

// RegisterStorePrefix allows other packages to register prefixes to be initalized on the store.
// The prefixes will be created on the store during GD2 context intialization
func RegisterStorePrefix(prefix string) {
	prefixes = append(prefixes, prefix)
}

// InitStore is to initialize the store
func InitStore(initPrefix bool) {
	Store = store.New()

	if initPrefix == true {

		if e := Store.InitPrefix(store.GlusterPrefix); e != nil {
			log.WithFields(log.Fields{
				"prefix": store.GlusterPrefix,
				"error":  e,
			}).Error("InitPrefix failed.")
		}

		for _, prefix := range prefixes {
			if e := Store.InitPrefix(prefix); e != nil {
				log.WithFields(log.Fields{
					"prefix": prefix,
					"error":  e,
				}).Error("InitPrefix failed.")
			}
		}
	}
}
