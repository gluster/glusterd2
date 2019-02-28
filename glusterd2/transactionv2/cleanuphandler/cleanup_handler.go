package cleanuphandler

import (
	"context"
	"expvar"
	"sync"
	"time"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transactionv2"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	log "github.com/sirupsen/logrus"
)

const (
	leaderKey       = "cleanup-leader"
	cleanupTimerDur = time.Minute * 5
	txnMaxAge       = time.Minute * 5
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
	isLeader   bool
	stopChan   chan struct{}
	stopOnce   sync.Once
	session    *concurrency.Session
	election   *concurrency.Election
	txnManager transaction.TxnManager
}

// WithSession configures a session with given ttl
func WithSession(client *clientv3.Client, ttl int) CleaupHandlerOptFunc {
	return func(handler *CleanupHandler) error {
		session, err := concurrency.NewSession(client, concurrency.WithTTL(ttl))
		if err != nil {
			return err
		}
		handler.session = session
		return nil
	}
}

// WithElection creates a new election for CleanupHandler.It will use the `defaultSession`
// if no session has been configured previously.
func WithElection(defaultSession *concurrency.Session) CleaupHandlerOptFunc {
	return func(handler *CleanupHandler) error {
		session := defaultSession
		if handler.session != nil {
			session = handler.session
		}
		handler.election = concurrency.NewElection(session, leaderKey)
		return nil
	}
}

// NewCleanupHandler returns a new CleanupHandler
func NewCleanupHandler(optFuncs ...CleaupHandlerOptFunc) (*CleanupHandler, error) {
	cl := &CleanupHandler{
		stopChan:   make(chan struct{}),
		txnManager: transaction.NewTxnManager(store.Store.Watcher),
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

// StartElecting triggers a new election campaign.
// If it succeeded then it assumes the leader role and returns
func (c *CleanupHandler) StartElecting() {
	log.Info("node started to contest for leader election")

	if err := c.election.Campaign(context.Background(), gdctx.MyUUID.String()); err != nil {
		log.WithError(err).Error("failed in campaign for cleanup leader election")
		c.Stop()
		return
	}

	log.Info("node got elected as cleanup leader")
	c.Lock()
	defer c.Unlock()
	events.Broadcast(newCleanupLeaderEvent())
	c.isLeader = true
}

// Stop will stop running the CleanupHandler
func (c *CleanupHandler) Stop() {
	log.Info("attempting to stop cleanup handler")
	c.stopOnce.Do(func() {
		close(c.stopChan)
		c.election.Resign(context.Background())
	})
}

// StartCleanupLeader starts cleanup leader
func StartCleanupLeader() {
	var err error

	CleanupLeader, err = NewCleanupHandler(
		WithSession(store.Store.NamespaceClient, 60),
		WithElection(store.Store.Session),
	)

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

func newCleanupLeaderEvent() *api.Event {
	data := map[string]string{
		"peer.id":   gdctx.MyUUID.String(),
		"peer.name": gdctx.HostName,
	}

	return events.New("cleanup leader elected", data, true)
}

func init() {
	expVar := expvar.Get("txn")
	if expVar == nil {
		expVar = expvar.NewMap("txn")
	}
	expVar.(*expvar.Map).Set("cleanup_config", expvar.Func(func() interface{} {
		return map[string]interface{}{
			"txn_max_age_seconds": txnMaxAge.Seconds(),
			"cleanup_dur_seconds": cleanupTimerDur.Seconds(),
		}
	}))
}
