package main

import (
	"io"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func initLog(logLevel string, out io.Writer) {
	l, err := log.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		log.WithField("error", err).Fatal("Failed to parse log level")
	}
	log.SetLevel(l)
	log.SetOutput(out)
}
