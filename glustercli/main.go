package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gluster/glusterd2/glustercli/cmd"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLogFile  = "./cli.log"
	defaultLogLevel = "INFO"
)

var logWriter io.WriteCloser

func openLogFile(filepath string) (io.WriteCloser, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func setLogOutput(w io.Writer) {
	log.SetOutput(w)
}

func initLog(logFilePath string, logLevel string) error {
	// Close the previously opened Log file
	if logWriter != nil {
		logWriter.Close()
		logWriter = nil
	}

	l, err := log.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		setLogOutput(os.Stderr)
		log.WithError(err).Debug("Failed to parse log level")
		return err
	}
	log.SetLevel(l)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	logFile, err := openLogFile(logFilePath)
	if err != nil {
		setLogOutput(os.Stderr)
		log.WithError(err).Debug("Failed to open log file %s", logFilePath)
		return err
	}
	setLogOutput(logFile)
	logWriter = logFile
	return nil
}
func main() {
	err := initLog(defaultLogFile, defaultLogLevel)
	if err != nil {
		fmt.Println("Error initializing log file ", err)
	}

	// Migrate old format Args into new Format. Modifies os.Args[]
	argsMigrate()

	if err = cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
