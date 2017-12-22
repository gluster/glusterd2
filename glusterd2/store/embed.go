package store

import (
	"path"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/elasticetcd"

	"github.com/coreos/etcd/pkg/types"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

func newEmbedStore(sconf *Config) (*GDStore, error) {
	econf, err := getElasticConfig(sconf)
	if err != nil {
		log.WithError(err).Error("failed to create embedded store config")
		return nil, err
	}

	log.WithFields(log.Fields{
		"name":          econf.Name,
		"datadir":       econf.Dir,
		"logdir":        econf.LogDir,
		"endpoints":     econf.Endpoints.String(),
		"curls":         econf.CURLs.String(),
		"purls":         econf.PURLs.String(),
		"certfile":      econf.CertFile,
		"keyfile":       econf.KeyFile,
		"cafile":        econf.CAFile,
		"trustedcafile": econf.TrustedCAFile,
	}).Debug("starting embedded store")

	ee, err := elasticetcd.New(econf)
	if err != nil {
		log.WithError(err).Error("failed to start embedded store")
		return nil, err
	}

	gds, err := newNamespacedStore(ee.Client(), sconf)
	if err != nil {
		return nil, err
	}

	gds.ee = ee

	return gds, nil
}

func (s *GDStore) closeEmbedStore() {
	log.Debug("stopping embedded store")
	s.ee.Stop()
	log.Debug("stopped embedded store")
}

func getElasticConfig(sconf *Config) (*elasticetcd.Config, error) {
	econf := elasticetcd.NewConfig()

	econf.Name = gdctx.MyUUID.String()
	econf.Dir = sconf.Dir
	econf.LogDir = path.Join(config.GetString("logdir"), "store")

	endpoints, err := types.NewURLs(sconf.Endpoints)
	if err != nil {
		return nil, err
	}
	curls, err := types.NewURLs(sconf.CURLs)
	if err != nil {
		return nil, err
	}
	purls, err := types.NewURLs(sconf.PURLs)
	if err != nil {
		return nil, err
	}
	econf.Endpoints = endpoints
	econf.CURLs = curls
	econf.PURLs = purls
	econf.UseTLS = sconf.UseTLS
	econf.CertFile = sconf.CertFile
	econf.KeyFile = sconf.KeyFile
	econf.CAFile = sconf.CAFile
	econf.TrustedCAFile = sconf.CAFile
	econf.ClntCertFile = sconf.ClntCertFile
	econf.ClntKeyFile = sconf.ClntKeyFile

	return econf, nil
}
