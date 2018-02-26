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
	// Prefix returns the prefix to be used for storing values
	Prefix() string
}

// Tctx represents structure for transaction context
type Tctx struct {
	parent    *Tctx
	log       log.FieldLogger // Functions which are given this context must use this logger to log their data.
	logFields log.Fields

	prefix string // The prefix under which the data is to be stored
}

// NewCtx returns a new empty TxnCtx with no parent, no associated data and the default logger.
func NewCtx() *Tctx {
	return &Tctx{
		log: log.StandardLogger(),
	}
}

// NewCtxWithLogFields returns a new context with the logger set to log given fields
func NewCtxWithLogFields(fields log.Fields) *Tctx {
	c := NewCtx()
	c.log = c.log.WithFields(fields)
	c.logFields = fields

	return c
}

// NewCtx returns a new empty TxnCtx with no parent, no associated data and the default logger.
func (c *Tctx) NewCtx() *Tctx {
	return &Tctx{
		parent:    c,
		log:       c.log,
		logFields: c.logFields,
		prefix:    c.prefix,
	}
}

// WithLogFields returns a new context with the logger set to log given fields
func (c *Tctx) WithLogFields(fields log.Fields) *Tctx {
	n := c.NewCtx()
	n.log = n.log.WithFields(fields)
	n.logFields = fields

	return n
}

// WithPrefix returns a new context with the store prefix set
func (c *Tctx) WithPrefix(prefix string) *Tctx {
	n := c.NewCtx()
	n.prefix = prefix

	return n
}

// Set attaches the given key-value pair to the context.
// If the key exists, the value will be updated.
func (c *Tctx) Set(key string, value interface{}) error {
	json, e := json.Marshal(value)
	if e != nil {
		c.log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to marshal value")
		return e
	}

	storeKey := c.prefix + "/" + key
	_, e = store.Store.Put(context.TODO(), storeKey, string(json))
	if e != nil {
		c.log.WithFields(log.Fields{
			"error": e,
			"key":   storeKey,
		}).Error("failed to set value")
	}
	return e
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

	storeKey := c.prefix + "/" + key
	r, err := store.Store.Get(context.TODO(), storeKey)
	if err != nil {
		c.log.WithError(err).WithField("key", storeKey).Error("failed to get value from store")
		return err
	}

	if r.Count == 0 {
		c.log.WithError(err).WithField("key", storeKey).Debug("key not found in store")
		return errors.New("key not found")
	}

	if err = json.Unmarshal(r.Kvs[0].Value, value); err != nil {
		c.log.WithError(err).WithField("key", storeKey).Error("failed to unmarshall value")
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
	storeKey := c.prefix + "/" + key
	_, e := store.Store.Delete(context.TODO(), storeKey)
	if e != nil {
		c.log.WithFields(log.Fields{
			"error": e,
			"key":   storeKey,
		}).Error("failed to delete key")
		return e
	}
	return nil
}

// Logger returns the Logrus logger associated with the context
func (c *Tctx) Logger() log.FieldLogger {
	return c.log
}

// Prefix returns the prefix to be used for storing values
func (c *Tctx) Prefix() string {
	return c.prefix
}

// Implementing the JSON Marshaler and Unmarshaler interfaces to allow Contexts
// to be exported Using an temporary struct to allow Context to be serialized
// using JSON.  Cannot serialize Context.Log otherwise.
// TODO: Implement proper tests to ensure proper Context is generated after (un)marshaling.
// XXX: We shold ideally be using protobuf here instead of JSON, as we use it for RPC,
// but JSON is simpler

type expContext struct {
	Parent    *Tctx
	LogFields log.Fields
	Prefix    string
}

// MarshalJSON implements the json.Marshaler interface
func (c *Tctx) MarshalJSON() ([]byte, error) {
	ac := expContext{
		Parent:    c.parent,
		LogFields: c.logFields,
		Prefix:    c.prefix,
	}

	return json.Marshal(ac)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *Tctx) UnmarshalJSON(d []byte) error {
	var ac expContext

	e := json.Unmarshal(d, &ac)
	if e != nil {
		return e
	}

	c.parent = ac.Parent
	if c.parent == nil {
		c.log = log.StandardLogger().WithFields(ac.LogFields)
	} else {
		c.log = c.parent.log.WithFields(ac.LogFields)
	}
	c.logFields = ac.LogFields
	c.prefix = ac.Prefix

	return nil
}
