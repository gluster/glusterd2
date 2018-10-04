package gdctx

import (
	"errors"
	"expvar"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/pborman/uuid"
	toml "github.com/pelletier/go-toml"
	config "github.com/spf13/viper"
)

const (
	uuidFileName    = "uuid.toml"
	peerIDKey       = "peer-id"
	clusterIDKey    = "cluster-id"
	envPrefix       = "GD2"
	envPeerIDKey    = envPrefix + "_PEER_ID"
	envClusterIDKey = envPrefix + "_CLUSTER_ID"
)

var (
	expPeerID    = expvar.NewString(peerIDKey)
	expClusterID = expvar.NewString(clusterIDKey)

	idMut sync.Mutex
)

var (
	// MyUUID is the unique identifier for this node in the cluster
	MyUUID uuid.UUID
	// MyClusterID is the unique identifier for the entire cluster
	MyClusterID uuid.UUID
)

func uuidFilePath() string {
	return path.Join(config.GetString("localstatedir"), uuidFileName)
}

// uuidConfig is a type that gives the configured values for peer and cluster ids
// from the following sources in order of preference
// - environment variables (GD2_CLUSTER_ID and GD2_PEER_ID)
// - the uuid config file ($LOCALSTATEDIR/uuid.toml)
// - randomly generated uuid
type uuidConfig struct {
	*config.Viper
}

func newUUIDConfig() *uuidConfig {
	uc := &uuidConfig{config.New()}

	// First setup config to use environment variables
	// TODO: Should be using a common prefix constant here and in the main
	// glusterd2 package
	uc.SetEnvPrefix(envPrefix)
	uc.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	uc.AutomaticEnv()

	// Next setup config to read in from the config file
	uc.SetConfigFile(uuidFilePath())
	uc.SetConfigType("toml")

	// Finally set random uuids as the default
	uc.SetDefault(peerIDKey, uuid.NewRandom().String())
	uc.SetDefault(clusterIDKey, uuid.NewRandom().String())

	return uc
}

func (uc *uuidConfig) reload(init bool) error {
	// Reload config from file
	if err := uc.ReadInConfig(); err != nil {
		// Error out if not initializing
		if !init {
			return err
		}

		// If initializing, ignore ENOENT error
		if !os.IsNotExist(err) {
			return err
		}
	}

	peerID := uc.GetString(peerIDKey)
	clusterID := uc.GetString(clusterIDKey)

	MyUUID = uuid.Parse(peerID)
	if MyUUID == nil {
		return errors.New("could not parse peer-id")
	}
	MyClusterID = uuid.Parse(clusterID)
	if MyClusterID == nil {
		return errors.New("could not parse cluster-id")
	}

	expPeerID.Set(MyUUID.String())
	expClusterID.Set(MyClusterID.String())

	return nil
}

func (uc *uuidConfig) save() error {
	tmpCfg := struct {
		PeerID    string `toml:"peer-id"`
		ClusterID string `toml:"cluster-id"`
	}{
		PeerID:    uc.GetString(peerIDKey),
		ClusterID: uc.GetString(clusterIDKey),
	}

	b, err := toml.Marshal(tmpCfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(uuidFilePath(), b, 0644)
}

// UpdateClusterID shall update the cluster ID and save it to file.
func UpdateClusterID(id string) error {
	idMut.Lock()
	defer idMut.Unlock()

	cfg := newUUIDConfig()
	cfg.Set(clusterIDKey, id)

	if err := cfg.save(); err != nil {
		return err
	}

	return cfg.reload(false)
}

// InitUUID intializes the peer and cluster IDs using the configured or saved
// values if available, or with random uuids
func InitUUID() error {
	idMut.Lock()
	defer idMut.Unlock()

	cfg := newUUIDConfig()
	if err := cfg.reload(true); err != nil {
		return err
	}

	return cfg.save()
}
