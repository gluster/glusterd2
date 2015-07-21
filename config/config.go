// Package config implements the runtime configuration of GlusterD
package config

type GDConfig struct {
	// TODO: Add things as required
	RestAddress string
}

func New() *GDConfig {
	return &GDConfig{}
}
