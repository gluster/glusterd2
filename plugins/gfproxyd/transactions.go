package gfproxyd

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
)

func txnGfproxydStart(c transaction.TxnCtx) error {
	var volname string
	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
		return err
	}

	gfproxyd, err := newgfproxyd(volname)
	if err != nil {
		return err
	}
	err = daemon.Start(gfproxyd, true)
	return err
}

func txnGfproxydStop(c transaction.TxnCtx) error {
	var volname string
	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
		return err
	}

	gfproxyd, err := newgfproxyd(volname)
	if err != nil {
		return err
	}
	err = daemon.Stop(gfproxyd, true)
	return err
}
