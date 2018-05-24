package transaction

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// TxnCtx is used to carry contextual information across the lifetime of a transaction
type TxnCtx interface {
	// Set attaches the given key with value to the context. It updates value if key exists already.
	Set(key string, value interface{}) error
	// SetNodeResult is similar to Set but prefixes the key with node UUID specified.
	SetNodeResult(nodeID uuid.UUID, key string, value interface{}) error
	// Get gets the value for the given key. Returns an error if the key is not present
	Get(key string, value interface{}) error
	// GetNodeResult is similar to Get but prefixes the key with node UUID specified.
	GetNodeResult(nodeID uuid.UUID, key string, value interface{}) error
	// Delete deletes the key and value
	Delete(key string) error
	// Logger returns the Logrus logger associated with the context
	Logger() log.FieldLogger
}

// Tctx represents structure for transaction context
type Tctx struct {
	logger log.FieldLogger
	config *txnCtxConfig
}

type txnCtxConfig struct {
	LogFields   log.Fields
	StorePrefix string
}

func newCtx(config *txnCtxConfig) *Tctx {
	return &Tctx{
		config: config,
		logger: log.StandardLogger().WithFields(config.LogFields),
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

	storeKey := c.config.StorePrefix + "/" + key
	if _, err = store.Store.Put(context.TODO(), storeKey, string(b)); err != nil {
		c.logger.WithError(err).WithField("key", key).Error("failed to set key in transaction context")
	}

	return err
}

// SetNodeResult is similar to Set but prefixes the key with the node UUID
// specified. This function can be used by nodes to store results of
// transaction steps.
func (c *Tctx) SetNodeResult(nodeID uuid.UUID, key string, value interface{}) error {
	storeKey := nodeID.String() + "/" + key
	return c.Set(storeKey, value)
}

// Get gets the value for the given key if available.
// Returns error if not found.
func (c *Tctx) Get(key string, value interface{}) error {

	storeKey := c.config.StorePrefix + "/" + key
	r, err := store.Store.Get(context.TODO(), storeKey)
	if err != nil {
		c.logger.WithError(err).WithField("key", storeKey).Error("failed to get value from transaction context")
		return err
	}

	if r.Count == 0 {
		c.logger.WithError(err).WithField("key", storeKey).Debug("key not found in store")
		return errors.New("key not found")
	}

	if err = json.Unmarshal(r.Kvs[0].Value, value); err != nil {
		c.logger.WithError(err).WithField("key", storeKey).Error("failed to unmarshall value")
	}

	return err
}

// GetNodeResult is similar to Get but prefixes the key with node UUID
// specified. This function can be used by the transaction initiator node to
// fetch results of transaction step run on remote nodes.
func (c *Tctx) GetNodeResult(nodeID uuid.UUID, key string, value interface{}) error {
	storeKey := nodeID.String() + "/" + key
	return c.Get(storeKey, value)
}

// Delete deletes the key and attached value
func (c *Tctx) Delete(key string) error {
	storeKey := c.config.StorePrefix + "/" + key
	if _, err := store.Store.Delete(context.TODO(), storeKey); err != nil {
		c.logger.WithError(err).WithField("key", storeKey).Error(
			"failed to delete key")
		return err
	}
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

	c.logger = log.StandardLogger().WithFields(c.config.LogFields)

	return nil
}
