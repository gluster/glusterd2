package cmd

import (
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gluster/glusterd2/pkg/restclient"
)

var client *restclient.Client
var logWriter io.WriteCloser

func initRESTClient() {
	client = restclient.New("http://localhost:24007", "", "")
}

func failure(msg string, err int) {
	os.Stderr.WriteString(msg + "\n")
	if err != 0 {
		os.Exit(err)
	}
}

func validateNArgs(cmd *cobra.Command, min int, max int) {
	nargs := len(cmd.Flags().Args())
	if nargs < min || (max != 0 && nargs > max) {
		cmd.Usage()
		os.Exit(1)
	}
}

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
