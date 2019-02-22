package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pborman/uuid"
)

// GlobalTxnManager stores and manages access to transaction related data
var GlobalTxnManager TxnManager

const (
	// TxnStatusPrefix is etcd key prefix under which status of a txn is stored for a particular node
	// eg.. key for storing status:-  pending-transaction/<txn-ID>/<node-ID>/status
	TxnStatusPrefix = "status"
	// LastExecutedStepPrefix is etcd key prefix under which last step executed on a particular node for a txn is stored
	// eg.. key for storing last executed step on a node:-  pending-transaction/<txn-ID>/<node-ID>/laststep
	LastExecutedStepPrefix = "laststep"
	// etcd txn timeout in seconds
	etcdTxnTimeout = time.Second * 60
)

// TxnManager stores and manages access to transaction related data in
// `pending-transaction` namespace.
type TxnManager interface {
	WatchTxn(stopCh <-chan struct{}) <-chan *Txn
	GetTxns() []*Txn
	AddTxn(txn *Txn) error
	GetTxnByUUID(id uuid.UUID) (*Txn, error)
	RemoveTransaction(txnID uuid.UUID) error
	UpdateLastExecutedStep(index int, txnID uuid.UUID, nodeIDs ...uuid.UUID) error
	GetLastExecutedStep(txnID uuid.UUID, nodeID uuid.UUID) (int, error)
	WatchLastExecutedStep(stopCh <-chan struct{}, txnID uuid.UUID, nodeID uuid.UUID) <-chan int
	WatchFailedTxn(stopCh <-chan struct{}, nodeID uuid.UUID) <-chan *Txn
	WatchTxnStatus(stopCh <-chan struct{}, txnID uuid.UUID, nodeID uuid.UUID) <-chan TxnStatus
	GetTxnStatus(txnID uuid.UUID, nodeID uuid.UUID) (TxnStatus, error)
	UpDateTxnStatus(state TxnStatus, txnID uuid.UUID, nodeIDs ...uuid.UUID) error
	TxnGC(maxAge time.Duration)
	RemoveFailedTxns()
}

type txnManager struct {
	sync.Mutex
	getStoreKey  func(...string) string
	storeWatcher clientv3.Watcher
}

// NewTxnManager returns a TxnManager
func NewTxnManager(storeWatcher clientv3.Watcher) TxnManager {
	tm := &txnManager{
		storeWatcher: storeWatcher,
	}
	tm.getStoreKey = func(s ...string) string {
		key := path.Join(PendingTxnPrefix, path.Join(s...))
		return key
	}
	return tm
}

// RemoveTransaction removes a transaction from `pending-transaction namespace`
func (tm *txnManager) RemoveTransaction(txnID uuid.UUID) error {
	_, err := store.Delete(context.TODO(), tm.getStoreKey(txnID.String()), clientv3.WithPrefix())
	return err
}

// WatchTxnStatus watches status of txn on a particular node
func (tm *txnManager) WatchTxnStatus(stopCh <-chan struct{}, txnID uuid.UUID, nodeID uuid.UUID) <-chan TxnStatus {
	var (
		txnStatusChan = make(chan TxnStatus, 10)
		key           = tm.getStoreKey(txnID.String(), nodeID.String(), TxnStatusPrefix)
	)

	respHandler := func(response clientv3.WatchResponse) {
		for _, event := range response.Events {
			txnStatus := TxnStatus{}
			if err := json.Unmarshal(event.Kv.Value, &txnStatus); err != nil {
				continue
			}
			if !txnStatus.State.Valid() {
				continue
			}
			txnStatusChan <- txnStatus
		}
	}

	tm.watch(stopCh, key, respHandler, clientv3.WithFilterDelete())
	return txnStatusChan
}

// WatchTxn watches for newly added txn to store
func (tm *txnManager) WatchTxn(stopCh <-chan struct{}) <-chan *Txn {
	var (
		txnChan = make(chan *Txn, 10)
		key     = tm.getStoreKey()
		opts    = []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithFilterDelete()}
	)

	respHandler := func(response clientv3.WatchResponse) {
		for _, txn := range tm.watchRespToTxns(response) {
			txnChan <- txn
		}
	}

	tm.watch(stopCh, key, respHandler, opts...)

	return txnChan
}

// WatchFailedTxn watches for a failed txn on a particular node
func (tm *txnManager) WatchFailedTxn(stopCh <-chan struct{}, nodeID uuid.UUID) <-chan *Txn {
	var (
		txnChan = make(chan *Txn)
		key     = tm.getStoreKey()
		ops     = []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithFilterDelete()}
	)

	go func() {
		resp, err := store.Get(context.TODO(), key, ops...)
		if err != nil {
			return
		}
		for _, kv := range resp.Kvs {
			if txn := tm.kvToFailedTxn(kv, nodeID); txn != nil {
				txnChan <- txn
			}
		}
	}()

	respHandler := func(resp clientv3.WatchResponse) {
		for _, event := range resp.Events {
			if txn := tm.kvToFailedTxn(event.Kv, nodeID); txn != nil {
				txnChan <- txn
			}
		}
	}

	tm.watch(stopCh, key, respHandler, ops...)
	return txnChan
}

func (tm *txnManager) kvToFailedTxn(kv *mvccpb.KeyValue, nodeID uuid.UUID) *Txn {

	if !strings.HasSuffix(string(kv.Key), TxnStatusPrefix) {
		return nil
	}

	prefix, _ := path.Split(string(kv.Key))
	nID := path.Base(prefix)

	if nodeID.String() != nID {
		return nil
	}

	txnStatus := &TxnStatus{}
	if err := json.Unmarshal(kv.Value, txnStatus); err != nil {
		return nil
	}

	if txnStatus.State != txnFailed {
		return nil
	}

	txn, err := tm.GetTxnByUUID(txnStatus.TxnID)
	if err != nil {
		return nil
	}
	return txn
}

func (tm *txnManager) watchRespToTxns(resp clientv3.WatchResponse) (txns []*Txn) {
	for _, event := range resp.Events {
		prefix, id := path.Split(string(event.Kv.Key))
		if uuid.Parse(id) == nil || prefix != PendingTxnPrefix {
			continue
		}

		txn := &Txn{Ctx: new(oldtransaction.Tctx)}
		if err := json.Unmarshal(event.Kv.Value, txn); err != nil {
			continue
		}

		txns = append(txns, txn)
	}
	return
}

// AddTxn adds a txn to the store
func (tm *txnManager) AddTxn(txn *Txn) error {
	data, err := json.Marshal(txn)
	if err != nil {
		return err
	}
	_, err = store.Put(context.TODO(), tm.getStoreKey(txn.ID.String()), string(data))
	return err
}

// GetTxnByUUID returns the txn from given ID
func (tm *txnManager) GetTxnByUUID(id uuid.UUID) (*Txn, error) {
	key := tm.getStoreKey(id.String())
	resp, err := store.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, errors.New(key + " key not found")
	}

	kv := resp.Kvs[0]

	txn := &Txn{Ctx: new(oldtransaction.Tctx)}
	if err := json.Unmarshal(kv.Value, txn); err != nil {
		return nil, err
	}
	return txn, nil
}

// GetTxns returns all txns added to the store
func (tm *txnManager) GetTxns() (txns []*Txn) {
	resp, err := store.Get(context.TODO(), tm.getStoreKey(), clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, kv := range resp.Kvs {
		prefix, id := path.Split(string(kv.Key))
		if uuid.Parse(id) == nil || prefix != PendingTxnPrefix {
			continue
		}

		txn := &Txn{Ctx: new(oldtransaction.Tctx)}
		if err := json.Unmarshal(kv.Value, txn); err != nil {
			continue
		}
		txns = append(txns, txn)
	}
	return
}

// UpDateTxnStatus updates txn status for given nodes
func (tm *txnManager) UpDateTxnStatus(status TxnStatus, txnID uuid.UUID, nodeIDs ...uuid.UUID) error {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), etcdTxnTimeout)
		clusterLock = oldtransaction.Locks{}
		putOps      []clientv3.Op
	)

	defer cancel()
	defer clusterLock.UnLock(ctx)

	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	for _, nodeID := range nodeIDs {
		key := tm.getStoreKey(txnID.String(), nodeID.String(), TxnStatusPrefix)
		if err := clusterLock.Lock(key); err != nil {
			return err
		}
		putOps = append(putOps, clientv3.OpPut(key, string(data)))
	}

	txn, err := store.Txn(ctx).Then(putOps...).Commit()
	if err != nil || !txn.Succeeded {
		return errors.New("etcd txn to update txn status failed")
	}
	return nil
}

// GetTxnStatus returns status of given txn on a particular node
func (tm *txnManager) GetTxnStatus(txnID uuid.UUID, nodeID uuid.UUID) (TxnStatus, error) {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		key         = tm.getStoreKey(txnID.String(), nodeID.String(), TxnStatusPrefix)
		clusterLock = oldtransaction.Locks{}
	)

	defer cancel()

	if err := clusterLock.Lock(key); err != nil {
		return TxnStatus{State: txnUnknown, Reason: err.Error()}, err
	}
	defer clusterLock.UnLock(ctx)

	resp, err := store.Get(context.TODO(), key)
	if err != nil {
		return TxnStatus{State: txnUnknown, Reason: err.Error()}, err
	}

	if len(resp.Kvs) == 0 {
		return TxnStatus{State: txnUnknown}, errors.New(key + " key not found")
	}

	txnStatus := TxnStatus{}
	kv := resp.Kvs[0]

	if err := json.Unmarshal(kv.Value, &txnStatus); err != nil {
		return TxnStatus{State: txnUnknown, Reason: err.Error()}, err
	}

	if !txnStatus.State.Valid() {
		return TxnStatus{State: txnUnknown}, errors.New("invalid txn state")
	}

	return txnStatus, nil
}

// UpdateLastExecutedStep updates the last executed step on a node of a given txn ID
func (tm *txnManager) UpdateLastExecutedStep(index int, txnID uuid.UUID, nodeIDs ...uuid.UUID) error {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), etcdTxnTimeout)
		putOps      []clientv3.Op
	)

	defer cancel()

	for _, nodeID := range nodeIDs {
		key := tm.getStoreKey(txnID.String(), nodeID.String(), LastExecutedStepPrefix)
		putOps = append(putOps, clientv3.OpPut(key, strconv.Itoa(index)))
	}

	txn, err := store.Txn(ctx).Then(putOps...).Commit()
	if err != nil || !txn.Succeeded {
		return errors.New("etcd txn to update last executed step failed")
	}
	return nil
}

// GetLastExecutedStep fetches the last executed step on a node for a given txn ID
func (tm *txnManager) GetLastExecutedStep(txnID uuid.UUID, nodeID uuid.UUID) (int, error) {
	key := tm.getStoreKey(txnID.String(), nodeID.String(), LastExecutedStepPrefix)

	resp, err := store.Get(context.TODO(), key)
	if err != nil {
		return -1, err
	}

	if resp.Count != 1 {
		return -1, errors.New("more than one entry for same key")
	}

	kv := resp.Kvs[0]
	return strconv.Atoi(string(kv.Value))
}

// WatchLastExecutedStep watches for last executed step on a node for a given txn ID
func (tm *txnManager) WatchLastExecutedStep(stopCh <-chan struct{}, txnID uuid.UUID, nodeID uuid.UUID) <-chan int {
	var (
		lastExecutedStepChan = make(chan int)
		key                  = tm.getStoreKey(txnID.String(), nodeID.String(), LastExecutedStepPrefix)
		opts                 = []clientv3.OpOption{clientv3.WithFilterDelete()}
	)

	resp, err := store.Get(context.TODO(), key)
	if err == nil && resp.Count == 1 {
		opts = append(opts, clientv3.WithRev(resp.Kvs[0].CreateRevision))
	}

	respHandler := func(response clientv3.WatchResponse) {
		for _, event := range response.Events {
			lastStep := string(event.Kv.Value)
			if i, err := strconv.Atoi(lastStep); err == nil {
				lastExecutedStepChan <- i
			}
		}
	}

	tm.watch(stopCh, key, respHandler, opts...)
	return lastExecutedStepChan
}

// TxnGC will mark all the expired txn as failed based on given maxAge
func (tm *txnManager) TxnGC(maxAge time.Duration) {
	tm.Lock()
	defer tm.Unlock()

	txns := tm.GetTxns()
	for _, txn := range txns {
		if txn.StartTime.Unix()+int64(maxAge/time.Second) < time.Now().Unix() {
			nonFailedNodes := []uuid.UUID{}
			for _, nodeID := range txn.Nodes {
				txnStatus, err := tm.GetTxnStatus(txn.ID, nodeID)
				if err == nil && txnStatus.State != txnFailed {
					nonFailedNodes = append(nonFailedNodes, nodeID)
				}
			}
			if len(nonFailedNodes) == 0 {
				continue
			}
			txnStatus := TxnStatus{State: txnFailed, TxnID: txn.ID, Reason: "txn expired"}
			txn.Ctx.Logger().Info("txn got expired marking as failure")
			tm.UpDateTxnStatus(txnStatus, txn.ID, nonFailedNodes...)
		}
	}
}

// RemoveFailedTxns will remove all failed txn if rollback is completed by all peers involved in the transaction.
func (tm *txnManager) RemoveFailedTxns() {
	txns := tm.GetTxns()

	for _, txn := range txns {
		nodesRollbacked := 0

		for _, nodeID := range txn.Nodes {
			txnStatus, err := tm.GetTxnStatus(txn.ID, nodeID)
			if err == nil && txnStatus.State == txnFailed {
				lastStep, err := tm.GetLastExecutedStep(txn.ID, nodeID)
				if err == nil && lastStep == -1 {
					nodesRollbacked++
				}
			}
		}

		if nodesRollbacked == len(txn.Nodes) {
			txn.Ctx.Logger().Info("txn rolled back on all nodes, cleaning from store")
			txn.removeContextData()
			tm.RemoveTransaction(txn.ID)
		}
	}
}

func (tm *txnManager) watch(stopCh <-chan struct{}, key string, respHandler func(clientv3.WatchResponse), opts ...clientv3.OpOption) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		watchRespChan := tm.storeWatcher.Watch(ctx, key, opts...)
		for {
			select {
			case <-stopCh:
				return
			case watchResp := <-watchRespChan:
				if watchResp.Err() != nil || watchResp.Canceled {
					return
				}
				respHandler(watchResp)

			}
		}
	}()
}
