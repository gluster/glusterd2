package blockprovider

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

// ProviderFunc returns a block Provider instance. It also returns an error
// If occurred any while creating a Provider instance
type ProviderFunc func() (Provider, error)

var (
	providersMutex  sync.Mutex
	providerFactory = make(map[string]ProviderFunc)
)

// Provider is an abstract, pluggable interface for block volume providers
type Provider interface {
	CreateBlockVolume(name string, size uint64, hostVolume string, options ...BlockVolOption) (BlockVolume, error)
	DeleteBlockVolume(name string, options ...BlockVolOption) error
	GetAndDeleteBlockVolume(name string, options ...BlockVolOption) (BlockVolume, error)
	GetBlockVolume(id string) (BlockVolume, error)
	BlockVolumes() []BlockVolume
	ProviderName() string
}

// BlockVolume is an interface which provides information about a block volume
type BlockVolume interface {
	Name() string
	ID() string
	HostVolume() string
	HostAddresses() []string
	IQN() string
	Username() string
	Password() string
	Size() uint64
	HaCount() int
}

// RegisterBlockProvider will register a block provider
func RegisterBlockProvider(name string, f ProviderFunc) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	if _, found := providerFactory[name]; found {
		log.WithField("name", name).Error("failed to register block provider, provider already exist")
		return
	}
	log.WithField("name", name).Debug("Registered block provider")
	providerFactory[name] = f
}

// GetBlockProvider will return a block Provider instance if it has been registered.
func GetBlockProvider(name string) (Provider, error) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	f, found := providerFactory[name]
	if !found {
		return nil, fmt.Errorf("%s block provider does not exist", name)
	}
	return f()
}
