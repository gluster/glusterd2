package elasticetcd

import (
	"errors"
)

var (
	// ErrClientNotAvailable is returned when the ElasticEtcd client is not present
	ErrClientNotAvailable = errors.New("etcd client not available")
	// ErrAddingSelfToServerList is returned when an ElasticEtcd instance fails to add itself to the nominated servers list
	ErrAddingSelfToServerList = errors.New("failed to add self to server list")
)
