package cleanuphandler

import (
	"context"
	"sync"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	txn "github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/transactionv2"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	leaderKey        = "leader"
	cleanupTimerDur  = time.Second * 5
	txnMaxAge        = time.Second * 20
	electionTimerDur = time.Second * 10
)

// CleanupLeader is responsible for performing all cleaning operation
var CleanupLeader *CleanupHandler

// CleaupHandlerOptFunc accepts a CleanupHandler and overrides its members
type CleaupHandlerOptFunc func(handler *CleanupHandler) error

// CleanupHandler performs all cleaning operation.
// It will remove all expired txn related data from store.
// A leader is elected among the peers in the cluster to
// cleanup stale transactions. The leader periodically scans
// the pending transaction namespace for failed and stale
// transactions, and cleans them up if rollback is completed
// by all peers involved in the transaction.
type CleanupHandler struct {
	sync.Mutex
	isLeader    bool
	locks       txn.Locks
	selfNodeID  uuid.UUID
	storeClient *clientv3.Client
	stopChan    chan struct{}
	stopOnce    sync.Once
	txnManager  transaction.TxnManager
}

// NewCleanupHandler returns a new CleanupHandler
func NewCleanupHandler(optFuncs ...CleaupHandlerOptFunc) (*CleanupHandler, error) {
	cl := &CleanupHandler{
		storeClient: store.Store.Client,
		stopChan:    make(chan struct{}),
		txnManager:  transaction.NewTxnManager(store.Store.Watcher),
		selfNodeID:  gdctx.MyUUID,
		locks:       make(txn.Locks),
	}

	for _, optFunc := range optFuncs {
		if err := optFunc(cl); err != nil {
			return nil, err
		}
	}

	return cl, nil
}

// Run starts running CleanupHandler
func (c *CleanupHandler) Run() {
	log.Info("cleanup handler started")

	go transaction.UntilStop(c.HandleStaleTxn, cleanupTimerDur, c.stopChan)
	go transaction.UntilStop(c.CleanFailedTxn, cleanupTimerDur, c.stopChan)

	<-c.stopChan
	log.Info("cleanup handler stopped")
}

// HandleStaleTxn will mark all the expired txn as failed based maxAge of a txn
func (c *CleanupHandler) HandleStaleTxn() {
	c.Lock()
	isLeader := c.isLeader
	c.Unlock()

	if isLeader {
		c.txnManager.TxnGC(txnMaxAge)
	}

}

// CleanFailedTxn removes all failed txn if rollback is
// completed by all peers involved in the transaction
func (c *CleanupHandler) CleanFailedTxn() {
	c.Lock()
	isLeader := c.isLeader
	c.Unlock()

	if isLeader {
		c.txnManager.RemoveFailedTxns()
	}
}

// Stop will stop running the CleanupHandler
func (c *CleanupHandler) Stop() {
	log.Info("attempting to stop cleanup handler")
	c.stopOnce.Do(func() {
		close(c.stopChan)
	})
	c.Lock()
	defer c.Unlock()
	isLeader := c.isLeader
	if isLeader {
		store.Delete(context.TODO(), leaderKey)
	}
	c.locks.UnLock(context.Background())
}

// StartElecting triggers a new election after every `electionTimerDur`.
// If it succeeded then it assumes the leader role and returns
func (c *CleanupHandler) StartElecting() {
	log.Info("node started to contest for leader election")

	transaction.UntilSuccess(c.IsNodeElected, electionTimerDur, c.stopChan)

	log.Info("node got elected as cleanup leader")
	c.Lock()
	defer c.Unlock()
	c.isLeader = true
}

// IsNodeElected returns whether a node is elected as a leader or not.
// Leader attempts to set a common key using a transaction that checks
// if the key already exists. If not, the candidate leader sets the
// key with a lease and assumes the leader role.
func (c *CleanupHandler) IsNodeElected() bool {
	var (
		leaseID = store.Store.Session.Lease()
		lockID  = gdctx.MyClusterID.String()
		logger  = log.WithField("lockID", lockID)
	)

	if err := c.locks.Lock(lockID); err != nil {
		logger.WithError(err).Error("error in acquiring lock")
		return false
	}
	defer c.locks.UnLock(context.Background())

	resp, err := store.Txn(context.Background()).
		If(clientv3.Compare(clientv3.Version(leaderKey), "=", 0)).
		Then(clientv3.OpPut(leaderKey, c.selfNodeID.String(), clientv3.WithLease(leaseID))).
		Commit()

	return (err == nil) && resp.Succeeded
}

// StartCleanupLeader starts cleanup leader
func StartCleanupLeader() {
	var err error

	CleanupLeader, err = NewCleanupHandler()

	if err != nil {
		log.WithError(err).Errorf("failed in starting cleanup handler")
		return
	}

	go CleanupLeader.StartElecting()
	go CleanupLeader.Run()
}

// StopCleanupLeader stops the cleanup leader
func StopCleanupLeader() {
	if CleanupLeader != nil {
		CleanupLeader.Stop()
	}
}
