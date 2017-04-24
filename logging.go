package main

import (
	"io"
	stdlog "log"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	stdlog.SetOutput(log.StandardLogger().Writer())
}

func initLog(logdir string, logFileName string, logLevel string) {
	// Close the previously opened Log file
	if logWriter != nil {
		logWriter.Close()
		logWriter = nil
	}

	l, err := log.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		log.SetOutput(os.Stderr)
		log.WithField("error", err).Fatal("Failed to parse log level")
	}
	log.SetLevel(l)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	if strings.ToLower(logFileName) == "stderr" || logFileName == "-" {
		setLogOutput(os.Stderr)
	} else if strings.ToLower(logFileName) == "stdout" {
		setLogOutput(os.Stdout)
	} else {
		logFilePath := path.Join(logdir, logFileName)
		logFile, logFileErr := openLogFile(logFilePath)
		if logFileErr != nil {
			setLogOutput(os.Stderr)
			log.WithError(logFileErr).Fatalf("Failed to open log file %s", logFilePath)
			return
		}
		setLogOutput(logFile)
		logWriter = logFile
	}
}
