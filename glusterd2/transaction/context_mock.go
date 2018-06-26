package transaction

import (
	"errors"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// MockTctx implements a dummy context type that can be used in tests
type MockTctx struct {
	data map[string]interface{}
}

// NewMockCtx returns a new instance of MockTctx
func NewMockCtx() *MockTctx {
	return &MockTctx{
		data: make(map[string]interface{}),
	}
}

// Set attaches the given key with value to the context. It updates value if key exists already.
func (m *MockTctx) Set(key string, value interface{}) error {
	m.data[key] = value
	return nil
}

// SetNodeResult is similar to Set but prefixes the key with the node UUID specified.
func (m *MockTctx) SetNodeResult(peerID uuid.UUID, key string, value interface{}) error {
	storeKey := peerID.String() + "/" + key
	return m.Set(storeKey, value)
}

// Get gets the value for the given key. Returns an error if the key is not present
func (m *MockTctx) Get(key string, value interface{}) error {
	_, ok := m.data[key]
	if !ok {
		return errors.New("key not present")
	}
	return nil
}

// GetNodeResult is similar to Get but prefixes the key with node UUID specified.
func (m *MockTctx) GetNodeResult(peerID uuid.UUID, key string, value interface{}) error {
	storeKey := peerID.String() + "/" + key
	return m.Get(storeKey, value)
}

// Delete deletes the key and value
func (m *MockTctx) Delete(key string) error {
	delete(m.data, key)
	return nil
}

// Logger returns a dummy logger
func (m *MockTctx) Logger() log.FieldLogger {
	return log.New()
}

// Prefix returns the prefix to be used for storing values
func (m MockTctx) Prefix() string {
	return "mock"
}
