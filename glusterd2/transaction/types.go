package transaction

import "github.com/pborman/uuid"

// TxnState represents state of a txn
type TxnState string

const (
	txnPending   TxnState = "Pending"
	txnRunning   TxnState = "Running"
	txnSucceeded TxnState = "Succeeded"
	txnFailed    TxnState = "Failed"
	txnUnknown   TxnState = "Unknown"
)

// Valid returns whether a TxnState is valid or not
func (ts TxnState) Valid() bool {
	switch ts {
	case txnPending:
		fallthrough
	case txnRunning:
		fallthrough
	case txnSucceeded:
		fallthrough
	case txnFailed:
		return true
	default:
		return false
	}
}

// TxnStatus represents status of a txn
type TxnStatus struct {
	State  TxnState  `json:"txn_state"`
	TxnID  uuid.UUID `json:"txn_id"`
	Reason string    `json:"reason,omitempty"`
}
