package transaction

import (
	"encoding/json"

	gdctx "github.com/gluster/glusterd2/context"

	log "github.com/Sirupsen/logrus"
)

// Context is used to carry contextual information across the lifetime of a request or a transaction.
type Context struct {
	parent *Context
	//data   map[string]interface{}
	Log    *log.Entry // Functions which are given this context must use this logger to log their data.
	Prefix string     // The prefix under which the data is to be stored
}

// NewEmptyContext returns a new empty Context with no parent, no associated data and the default logger.
func NewEmptyContext() *Context {
	return &Context{
		Log: log.NewEntry(log.StandardLogger()), //empty logging context
	}
}

// NewLoggingContext returns a new context with the logger set to log given fields
func NewLoggingContext(fields log.Fields) *Context {
	c := NewEmptyContext()
	c.Log = c.Log.WithFields(fields)

	return c
}

// NewContext returns a new empty context with given parent
func (c *Context) NewContext() *Context {
	return &Context{
		parent: c,
		Log:    c.Log,
		Prefix: c.Prefix,
	}
}

// NewLoggingContext returns a new context with the logger set to log given fields in addition to the parents logging fields
func (c *Context) NewLoggingContext(fields log.Fields) *Context {
	n := c.NewContext()
	n.Log = n.Log.WithFields(fields)

	return n
}

// Set attaches the given key-value pair to the context.
// If the key exists, the value will be updated.
func (c *Context) Set(key string, value interface{}) error {
	json, e := json.Marshal(value)
	if e != nil {
		c.Log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to marshal value")
		return e
	}
	e = gdctx.Store.Put(c.Prefix+key, json, nil)
	if e != nil {
		c.Log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to set value")
	}
	return e
}

// Get gets the value for the given key if available.
// Get recursively searches all parent contexts for the key.
// Returns nil if not found.
func (c *Context) Get(key string, value interface{}) error {
	b, e := gdctx.Store.Get(c.Prefix + key)
	if e != nil {
		c.Log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to get value")
		return e
	}

	if e = json.Unmarshal(b.Value, value); e != nil {
		c.Log.WithFields(log.Fields{
			"error": e,
			"key":   key,
		}).Error("failed to unmarshal value")
	}
	return e
}

// Delete deletes the key and attached value
// Delete doesn't recurse to parents
func (c *Context) Delete(key string) error {
	return gdctx.Store.Delete(c.Prefix + key)
}

// Implementing the JSON Marshaler and Unmarshaler interfaces to allow Contexts
// to be exported Using an temporary struct to allow Context to be serialized
// using JSON.  Cannot serialize Context.Log otherwise.
// TODO: Implement proper tests to ensure proper Context is generated after (un)marshaling.
// XXX: We shold ideally be using protobuf here instead of JSON, as we use it for RPC,
// but JSON is simpler

type expContext struct {
	Parent *Context
	//Data      map[string]interface{}
	LogFields log.Fields
	Prefix    string
}

// MarshalJSON implements the json.Marshaler interface
func (c *Context) MarshalJSON() ([]byte, error) {
	ac := expContext{
		Parent: c.parent,
		//Data:      c.data,
		LogFields: c.Log.Data,
		Prefix:    c.Prefix,
	}

	return json.Marshal(ac)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *Context) UnmarshalJSON(d []byte) error {
	var ac expContext

	e := json.Unmarshal(d, &ac)
	if e != nil {
		return e
	}

	c.parent = ac.Parent
	//c.data = ac.Data
	if c.parent == nil {
		c.Log = log.NewEntry(log.StandardLogger()).WithFields(ac.LogFields)
	} else {
		c.Log = c.parent.Log.WithFields(ac.LogFields)
	}
	c.Prefix = ac.Prefix

	return nil
}
