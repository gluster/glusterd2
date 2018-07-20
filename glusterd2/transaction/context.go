package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// TxnCtx is used to carry contextual information across the lifetime of a transaction
type TxnCtx interface {
	// Set attaches the given key with value to the context. It updates value if key exists already.
	Set(key string, value interface{}) error
	// SetNodeResult is similar to Set but prefixes the key with node UUID specified.
	SetNodeResult(peerID uuid.UUID, key string, value interface{}) error
	// Get gets the value for the given key. Returns an error if the key is not present
	Get(key string, value interface{}) error
	// GetNodeResult is similar to Get but prefixes the key with node UUID specified.
	GetNodeResult(peerID uuid.UUID, key string, value interface{}) error
	// GetTxnReqID gets the reqID string saved in the transaction.
	GetTxnReqID() string
	// Delete deletes the key and value
	Delete(key string) error
	// Logger returns the Logrus logger associated with the context
	Logger() log.FieldLogger

	// commit writes all locally cached keys and values into the store using
	// a single etcd transaction. This is for internal use by the txn framework
	// and hence isn't exported.
	commit() error
}

// Tctx represents structure for transaction context
type Tctx struct {
	config         *txnCtxConfig // this will be marshalled and sent on wire
	logger         log.FieldLogger
	readSet        map[string][]byte // cached responses from store
	readCacheDirty bool
	writeSet       map[string]string // to be written to store
}

// txnCtxConfig is marshalled and sent on wire and is used to reconstruct Tctx
// on receiver's end.
type txnCtxConfig struct {
	LogFields   log.Fields
	StorePrefix string
}

func newCtx(config *txnCtxConfig) *Tctx {
	return &Tctx{
		config:         config,
		logger:         log.StandardLogger().WithFields(config.LogFields),
		readSet:        make(map[string][]byte),
		writeSet:       make(map[string]string),
		readCacheDirty: true,
	}
}

// Set attaches the given key-value pair to the context.
// If the key exists, the value will be updated.
func (c *Tctx) Set(key string, value interface{}) error {

	b, err := json.Marshal(value)
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("failed to marshal value")
		return err
	}

	storeKey := c.config.StorePrefix + key

	// Update the read cache to serve future local Get()s for this key from cache
	c.readSet[storeKey] = b

	// Update write cache, the contents of which will be committed to store later
	c.writeSet[storeKey] = string(b)

	return nil
}

// commit writes all locally cached keys and values into the store using
// a single etcd transaction.
func (c *Tctx) commit() error {

	if len(c.writeSet) == 0 {
		return nil
	}

	var putOps []clientv3.Op
	for key, value := range c.writeSet {
		putOps = append(putOps, clientv3.OpPut(key, value))
	}

	txn, err := store.Store.Txn(context.TODO()).
		If().
		Then(putOps...).
		Else().
		Commit()

	if err != nil || !txn.Succeeded {
		msg := "etcd txn to store txn context keys failed"
		if err == nil {
			// if txn.Succeeded = false
			err = errors.New(msg)
		}
		c.logger.WithError(err).WithField("keys",
			reflect.ValueOf(c.writeSet).MapKeys()).Error(msg)
		return err
	}

	expTxn.Add("txn_ctx_store_commit", 1)

	c.readCacheDirty = true

	return nil
}

// SetNodeResult is similar to Set but prefixes the key with the node UUID
// specified. This function can be used by nodes to store results of
// transaction steps.
func (c *Tctx) SetNodeResult(peerID uuid.UUID, key string, value interface{}) error {
	storeKey := peerID.String() + "/" + key
	return c.Set(storeKey, value)
}

// Get gets the value for the given key if available.
// Returns error if not found.
func (c *Tctx) Get(key string, value interface{}) error {

	// cache all keys and values from the store on the first call to Get
	if c.readCacheDirty {
		resp, err := store.Get(context.TODO(), c.config.StorePrefix, clientv3.WithPrefix())
		if err != nil {
			c.logger.WithError(err).WithField("key", key).Error("failed to get key from transaction context")
			return err
		}
		expTxn.Add("txn_ctx_store_get", 1)
		for _, kv := range resp.Kvs {
			c.readSet[string(kv.Key)] = kv.Value
		}
		c.readCacheDirty = false
	}

	// return cached key
	storeKey := c.config.StorePrefix + key
	if data, ok := c.readSet[storeKey]; ok {
		if err := json.Unmarshal(data, value); err != nil {
			c.logger.WithError(err).WithField("key", storeKey).Error("failed to unmarshall value")
		}
	} else {
		return errors.New("key not found")
	}

	return nil
}

// GetNodeResult is similar to Get but prefixes the key with node UUID
// specified. This function can be used by the transaction initiator node to
// fetch results of transaction step run on remote nodes.
func (c *Tctx) GetNodeResult(peerID uuid.UUID, key string, value interface{}) error {
	storeKey := peerID.String() + "/" + key
	return c.Get(storeKey, value)
}

// GetTxnReqID gets the reqID string saved within the txnCtxConfig.
func (c *Tctx) GetTxnReqID() string {
	return c.config.LogFields["reqid"].(string)
}

// Delete deletes the key and attached value
func (c *Tctx) Delete(key string) error {

	storeKey := c.config.StorePrefix + key

	delete(c.readSet, storeKey)
	delete(c.writeSet, storeKey)

	// TODO: Optimize this by doing it as part of etcd txn in commit()
	if _, err := store.Delete(context.TODO(), storeKey); err != nil {
		c.logger.WithError(err).WithField("key", storeKey).Error(
			"failed to delete key")
		return err
	}
	expTxn.Add("txn_ctx_store_delete", 1)
	return nil
}

// Logger returns the Logrus logger associated with the context
func (c *Tctx) Logger() log.FieldLogger {
	return c.logger
}

// MarshalJSON implements the json.Marshaler interface
func (c *Tctx) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.config)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *Tctx) UnmarshalJSON(d []byte) error {

	if err := json.Unmarshal(d, &c.config); err != nil {
		return err
	}

	*c = *(newCtx(c.config))

	return nil
}
