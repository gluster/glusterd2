package gdctx

import (
	"expvar"
	"io/ioutil"
	"os"
	"path"

	"github.com/pborman/uuid"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

var (
	expPeerID    = expvar.NewString("peer-id")
	expClusterID = expvar.NewString("cluster-id")
)

var (
	// MyUUID is the unique identifier for this node in the cluster
	MyUUID uuid.UUID
	// MyClusterID is the unique identifier for the entire cluster
	MyClusterID uuid.UUID
)

const uuidFileName = "uuid.toml"

// UUIDConfig is a type that is read from and written to uuidFileName file.
type UUIDConfig struct {
	PeerID    string `toml:"peer-id"`
	ClusterID string `toml:"cluster-id"`
}

func (cfg *UUIDConfig) reload() error {

	uuidFilePath := path.Join(config.GetString("localstatedir"), uuidFileName)
	b, err := ioutil.ReadFile(uuidFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		if err := toml.Unmarshal(b, cfg); err != nil {
			return err
		}
	}

	if cfg.PeerID == "" {
		cfg.PeerID = uuid.New()
		log.WithField("peer-id", cfg.PeerID).Info("Generated new peer ID")
	}

	if cfg.ClusterID == "" {
		cfg.ClusterID = uuid.New()
		log.WithField("cluster-id", cfg.ClusterID).Info("Generated new cluster ID")
	}

	MyUUID = uuid.Parse(cfg.PeerID)
	MyClusterID = uuid.Parse(cfg.ClusterID)
	expPeerID.Set(MyUUID.String())
	expClusterID.Set(MyClusterID.String())

	return nil
}

func (cfg *UUIDConfig) save() error {

	b, err := toml.Marshal(*cfg)
	if err != nil {
		return err
	}

	uuidFilePath := path.Join(config.GetString("localstatedir"), uuidFileName)
	return ioutil.WriteFile(uuidFilePath, b, 0644)
}

// UpdateClusterID shall update the cluster ID and save it to file.
func UpdateClusterID(id string) error {
	cfg := &UUIDConfig{
		PeerID:    MyUUID.String(),
		ClusterID: id,
	}

	if err := cfg.save(); err != nil {
		return err
	}

	return cfg.reload()
}

// InitUUID will generate (or use if present) node ID and cluster ID.
func InitUUID() error {
	cfg := &UUIDConfig{}

	if err := cfg.reload(); err != nil {
		return err
	}

	return cfg.save()
}
