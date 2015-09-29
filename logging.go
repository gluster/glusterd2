package main

import (
	"os"
	"strings"

	"github.com/gluster/glusterd2/config"

	log "github.com/Sirupsen/logrus"
)

func init() {
	l, err := log.ParseLevel(strings.ToLower(*config.LogLevel))
	if err != nil {
		log.WithField("error", err).Fatal("Failed to parse log level")
	}
	log.SetLevel(l)
	log.SetOutput(os.Stderr)
}
