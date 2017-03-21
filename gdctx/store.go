package gdctx

import (
	"github.com/gluster/glusterd2/store"
)

// If someone needs to use the GD2 store, all they need to do is just import context and use context.Store
var (
	Store    *store.GDStore
	prefixes []string
)

// InitStore is to initialize the store
func InitStore(initPrefix bool) {
	Store = store.New()
}
