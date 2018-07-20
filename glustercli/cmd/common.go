package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gluster/glusterd2/pkg/restclient"
)

var (
	client                       *restclient.Client
	logWriter                    io.WriteCloser
	errFailedToConnectToGlusterd = `Failed to connect to glusterd. Please check if
- Glusterd is running(%s) and reachable from this node.
- Make sure Endpoints specified in the command is valid
`
)

func initRESTClient(hostname, user, secret, cacert string, insecure bool) {
	client = restclient.New(hostname, user, secret, cacert, insecure)
	client.SetTimeout(time.Duration(flagTimeout) * time.Second)
}

func isConnectionRefusedErr(err error) bool {
	return strings.Contains(err.Error(), "connection refused")
}

func isNoSuchHostErr(err error) bool {
	return strings.Contains(err.Error(), "no such host")
}

func isNoRouteToHostErr(err error) bool {
	return strings.Contains(err.Error(), "no route to host")
}

func handleGlusterdConnectFailure(msg, endpoints string, err error, errcode int) {
	if err == nil {
		return
	}

	if isConnectionRefusedErr(err) || isNoSuchHostErr(err) || isNoRouteToHostErr(err) {
		os.Stderr.WriteString(msg + "\n\n")
		os.Stderr.WriteString(fmt.Sprintf(errFailedToConnectToGlusterd, endpoints))
		os.Exit(errcode)
	}
}

func failure(msg string, err error, errcode int) {
	handleGlusterdConnectFailure(msg, flagEndpoints[0], err, errcode)

	// If different error
	os.Stderr.WriteString(msg + "\n")
	if err != nil {
		os.Stderr.WriteString("\nError: " + err.Error() + "\n")
	}
	os.Exit(errcode)
}
