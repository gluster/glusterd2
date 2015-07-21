package store

import (
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
)

const (
	glusterPrefix string = "gluster/"
)

type Store struct {
	store.Store
}

func New() *Store {
	address := "localhost:8500"

	s, err := libkv.NewStore(store.CONSUL, []string{address}, &store.Config{ConnectionTimeout: 10 * time.Second})
	if err != nil {
		return nil
	}
	return &Store{s}
}
