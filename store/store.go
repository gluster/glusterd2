// Package store implements the centralized store for GlusterD
//
// It currently uses [Consul](https://www.consul.io) as the backend. But we
// hope to have it pluggable in the end. libkv is being used to help achieve
// this goal.
package store

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
)

const (
	glusterPrefix string = "gluster/"
)

type GDStore struct {
	store.Store
}

func New() *GDStore {
	//TODO: Make this configurable
	address := "localhost:8500"

	log.WithFields(log.Fields{"type": "consul", "consul.config": address}).Debug("Creating new store")
	s, err := libkv.NewStore(store.CONSUL, []string{address}, &store.Config{ConnectionTimeout: 10 * time.Second})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create store")
	}

	log.Info("Created new store using Consul")

	return &GDStore{s}
}
