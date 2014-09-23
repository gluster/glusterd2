package consul

import (
	"github.com/armon/consul-api"
)

const (
	glusterPrefix string = "gluster/"
)

type Consul struct {
	client *consulapi.Client
	kv     *consulapi.KV
}

func New() *Consul {
	c := new(Consul)
	c.init()

	return c
}

func (c *Consul) init() {
	c.client, _ = consulapi.NewClient(consulapi.DefaultConfig())
	c.kv = c.client.KV()
}
