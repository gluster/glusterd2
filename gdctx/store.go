package gdctx

import "github.com/gluster/glusterd2/store"

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
func initStore() {
	Store = store.New(Restart)

	// If its a fresh install and GlusterD is coming up for the first time
	// then initialize the store prefix, otherwise not
	if Restart == false {
		for _, prefix := range prefixes {
			Store.InitPrefix(prefix)
		}
	}
}
