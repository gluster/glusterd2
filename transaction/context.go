package transaction

import (
	"encoding/json"

	"github.com/gluster/glusterd2/gdctx"

	log "github.com/Sirupsen/logrus"
)

// TxnCtx is used to carry contextual information across the lifetime of a transaction
type TxnCtx interface {
	// Set attaches the given key with value to the context. It updates value if key exists already.
	Set(key string, value interface{}) error
	// Get gets the value for the given key. Returns an error if the key is not present
	Get(key string, value interface{}) error
	// Delete deletes the key and value
	Delete(key string) error
	// Logger returns the Logrus logger associated with the context
	Logger() log.FieldLogger
	// Prefix returns the prefix to be used for storing values
	Prefix() string
}

type txnCtx struct {
	parent *txnCtx
	//data   map[string]interface{}
	log       log.FieldLogger // Functions which are given this context must use this logger to log their data.
	logFields log.Fields

	prefix string // The prefix under which the data is to be stored
}

// NewCtx returns a new empty TxnCtx with no parent, no associated data and the default logger.
func NewCtx() *txnCtx {
	return &txnCtx{
		log: log.StandardLogger(),
	}
}

// NewCtxWithLogFields returns a new context with the logger set to log given fields
func NewCtxWithLogFields(fields log.Fields) *txnCtx {
	c := NewCtx()
	c.log = c.log.WithFields(fields)
	c.logFields = fields

	return c
}

// NewTxnCtxWithPrefix returns a new context with the store prefix set
func NewCtxWithPrefix(prefix string) *txnCtx {
	c := NewCtx()
	c.prefix = prefix

	return c
}

// NewTxnCtx returns a new empty TxnCtx with no parent, no associated data and the default logger.
func (c *txnCtx) NewCtx() *txnCtx {
	return &txnCtx{
		parent:    c,
		log:       c.log,
		logFields: c.logFields,
		prefix:    c.prefix,
	}
}

// WithLogFields returns a new context with the logger set to log given fields
func (c *txnCtx) WithLogFields(fields log.Fields) *txnCtx {
	n := c.NewCtx()
	n.log = n.log.WithFields(fields)
	n.logFields = fields

	return n
}

// WithPrefix returns a new context with the store prefix set
func (c *txnCtx) WithPrefix(prefix string) *txnCtx {
	n := c.NewCtx()
	n.prefix = prefix

	return n
}

// Set attaches the given key-value pair to the context.
// If the key exists, the value will be updated.
func (c *txnCtx) Set(key string, value interface{}) error {
	json, e := json.Marshal(value)
	if e != nil {
		c.log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to marshal value")
		return e
	}
	e = gdctx.Store.Put(c.prefix+key, json, nil)
	if e != nil {
		c.log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to set value")
	}
	return e
}

// Get gets the value for the given key if available.
// Returns error if not found.
func (c *txnCtx) Get(key string, value interface{}) error {
	b, e := gdctx.Store.Get(c.prefix + key)
	if e != nil {
		c.log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to get value")
		return e
	}

	if e = json.Unmarshal(b.Value, value); e != nil {
		c.log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to unmarshal value")
	}
	return e
}

// Delete deletes the key and attached value
func (c *txnCtx) Delete(key string) error {
	return gdctx.Store.Delete(c.prefix + key)
}

// Logger returns the Logrus logger associated with the context
func (c *txnCtx) Logger() log.FieldLogger {
	return c.log
}

// Prefix returns the prefix to be used for storing values
func (c *txnCtx) Prefix() string {
	return c.prefix
}

// Implementing the JSON Marshaler and Unmarshaler interfaces to allow Contexts
// to be exported Using an temporary struct to allow Context to be serialized
// using JSON.  Cannot serialize Context.Log otherwise.
// TODO: Implement proper tests to ensure proper Context is generated after (un)marshaling.
// XXX: We shold ideally be using protobuf here instead of JSON, as we use it for RPC,
// but JSON is simpler

type expContext struct {
	Parent    *txnCtx
	LogFields log.Fields
	Prefix    string
}

// MarshalJSON implements the json.Marshaler interface
func (c *txnCtx) MarshalJSON() ([]byte, error) {
	ac := expContext{
		Parent:    c.parent,
		LogFields: c.logFields,
		Prefix:    c.prefix,
	}

	return json.Marshal(ac)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *txnCtx) UnmarshalJSON(d []byte) error {
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
