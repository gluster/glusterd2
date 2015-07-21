// Package context is the runtime context of GlusterD
package context

import (
	"github.com/kshlm/glusterd2/config"
	"github.com/kshlm/glusterd2/rest"
	"github.com/kshlm/glusterd2/transaction"

	"github.com/Sirupsen/logrus"
)

type GDContext struct {
	Config *config.GDConfig
	Rest   *rest.GDRest
	TxnFw  *transaction.GDTxnFw
	Log    *logrus.Logger
}

func New() *GDContext {
	ctx := &GDContext{}

	return ctx
}
