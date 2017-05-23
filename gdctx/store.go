package gdctx

import (
	"github.com/gluster/glusterd2/store"
)

// Store variable can be imported by packages which need access to the store
var Store *store.GDStore

// InitStore is to initialize the store
func InitStore() error {
	var err error
	Store, err = store.New()
	return err
}
