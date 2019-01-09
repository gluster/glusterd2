package glusterblock

import (
	"github.com/spf13/viper"
)

// ClientConfig holds various config information needed to create a gluster-block rest client
type ClientConfig struct {
	HostAddress string
	User        string
	Secret      string
	CaCertFile  string
	Insecure    bool
}

// ApplyFromConfig sets the ClientConfig options from various config sources
func (c *ClientConfig) ApplyFromConfig(conf *viper.Viper) {
	c.CaCertFile = conf.GetString("gluster-block-cacert")
	c.HostAddress = conf.GetString("gluster-block-hostaddr")
	c.User = conf.GetString("gluster-block-user")
	c.Secret = conf.GetString("gluster-block-secret")
	c.Insecure = conf.GetBool("gluster-block-insecure")
}
