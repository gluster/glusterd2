package context

import (
	log "github.com/Sirupsen/logrus"
)

// Context is used to carry contextual information across the lifetime of a request or a transaction.
type Context struct {
	parent *Context
	data   map[string]interface{}

	Log *log.Entry // Functions which are given this context must use this logger to log their data.
}

// NewEmptyContext returns a new empty Context with no parent, no associated data and the default logger.
func NewEmptyContext() *Context {
	return &Context{
		data: make(map[string]interface{}),
		Log:  log.NewEntry(log.StandardLogger()), //empty logging context
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
		data:   make(map[string]interface{}),
		Log:    c.Log,
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
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get gets the value for the given key if available.
// Get recursively searches all parent contexts for the key.
// Returns nil if not found.
func (c *Context) Get(key string) interface{} {
	if c.data[key] != nil {
		return c.data[key]
	}
	return c.parent.Get(key)
}

// Delete deletes the key and attached value
// Delete doesn't recurse to parents
func (c *Context) Delete(key string) {
	delete(c.data, key)
}
