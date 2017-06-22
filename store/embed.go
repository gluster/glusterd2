package store

import (
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/elasticetcd"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/pkg/types"
)

func newEmbedStore(sconf *Config) (*GDStore, error) {
	econf, err := getElasticConfig(sconf)
	if err != nil {
		log.WithError(err).Error("failed to create embedded store config")
		return nil, err
	}

	log.WithFields(log.Fields{
		"name":      econf.Name,
		"datadir":   econf.Dir,
		"endpoints": econf.Endpoints.String(),
		"curls":     econf.CURLs.String(),
		"purls":     econf.PURLs.String(),
	}).Debug("starting embedded store")

	ee, err := elasticetcd.New(econf)
	if err != nil {
		log.WithError(err).Error("failed to start embedded store")
		return nil, err
	}

	return &GDStore{*sconf, ee.Client(), ee.Session(), ee}, nil
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

	return econf, nil
}
