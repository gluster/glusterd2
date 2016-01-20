package context

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

func initStore() {
	Store = store.New()

	for _, prefix := range prefixes {
		Store.InitPrefix(prefix)
	}
}
