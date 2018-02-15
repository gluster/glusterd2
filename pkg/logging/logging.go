// Package logging implements a common log intialization for GD2 and its CLI
package logging

import (
	"io"
	stdlog "log"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// DirFlag is the common logging flag to be used to set log directory
	DirFlag = "logdir"
	// DirHelp is the help message for DirFlag
	DirHelp = "Directory to store log files"

	// FileFlag is the common logging flag to be used to set log file name
	FileFlag = "logfile"
	// FileHelp is the help message for FileFlag
	FileHelp = "Name for log file"

	// LevelFlag is the common logging flag to be used to set log level
	LevelFlag = "loglevel"
	// LevelHelp is the help message for LevelFlag
	LevelHelp = "Severity of messages to be logged"

	// YY-MM-DD HH:MM:SS.SSSSSS
	timestampFormat = "2006-01-02 15:04:05.000000"
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

// Init initializes the default logrus logger
// Should be called as early as possible when a process starts.
// Note that this does not create a new logger. Packages should still continue
// importing and using logrus as before.
func Init(logdir string, logFileName string, logLevel string, verboseLogEntry bool) error {

	if verboseLogEntry {
		// TODO: Make it configurable with default being off. This
		// has performance overhead as it allocates memory every time.
		log.AddHook(SourceLocationHook{})
	}

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
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: timestampFormat})

	if strings.ToLower(logFileName) == "stderr" || logFileName == "-" {
		setLogOutput(os.Stderr)
	} else if strings.ToLower(logFileName) == "stdout" {
		setLogOutput(os.Stdout)
	} else {
		logFilePath := path.Join(logdir, logFileName)
		logFile, err := openLogFile(logFilePath)
		if err != nil {
			setLogOutput(os.Stderr)
			log.WithError(err).Debug("Failed to open log file %s", logFilePath)
			return err
		}
		setLogOutput(logFile)
		logWriter = logFile
	}
	return nil
}
