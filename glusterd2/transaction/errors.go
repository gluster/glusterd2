package transaction

import "errors"

var (
	errTxnTimeout     = errors.New("txn timeout")
	errTxnSyncTimeout = errors.New("timeout in synchronizing txn")
)
