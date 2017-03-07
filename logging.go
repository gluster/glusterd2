package main

import (
	"io"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var logWriter io.WriteCloser

func reloadLog(logdir string, logFileName string, logLevel string) {
	if logFileName != "-" {
		// Close the previously opened Log file
		if logWriter != nil {
			logWriter.Close()
		}

		// Reopen the log file again and update logWriter
		logFilePath := path.Join(logdir, logFileName)
		logFile, logFileErr := openLogFile(logFilePath)
		if logFileErr != nil {
			initLog(logLevel, os.Stderr)
			log.WithError(logFileErr).Fatalf("Failed to open log file %s", logFilePath)
			return
		}

		log.SetOutput(logFile)
		logWriter = logFile
	}
}

func openLogFile(filepath string) (io.WriteCloser, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func initLog(logLevel string, out io.WriteCloser) {
	l, err := log.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		log.WithField("error", err).Fatal("Failed to parse log level")
	}
	log.SetLevel(l)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(out)
	logWriter = out
}
