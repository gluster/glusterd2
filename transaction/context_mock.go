package transaction

import (
	"errors"
	log "github.com/Sirupsen/logrus"
)

// MockCtx implements a dummy context type that can be used in tests
type mockTxnCtx struct {
	data map[string]interface{}
}

func NewMockCtx() *mockTxnCtx {
	return &mockTxnCtx{
		data: make(map[string]interface{}),
	}
}

// Set attaches the given key with value to the context. It updates value if key exists already.
func (m *mockTxnCtx) Set(key string, value interface{}) error {
	m.data[key] = value
	return nil
}

// Get gets the value for the given key. Returns an error if the key is not present
func (m *mockTxnCtx) Get(key string, value interface{}) error {
	v, ok := m.data[key]
	if !ok {
		return errors.New("key not present")
	}
	value = v
	return nil
}

// Delete deletes the key and value
func (m *mockTxnCtx) Delete(key string) error {
	delete(m.data, key)
	return nil
}

// Logger returns a dummy logger
func (m *mockTxnCtx) Logger() log.FieldLogger {
	return log.New()
}

// Prefix returns the prefix to be used for storing values
func (m mockTxnCtx) Prefix() string {
	return "mock"
}
