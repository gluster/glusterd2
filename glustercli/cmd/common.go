package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gluster/glusterd2/pkg/restclient"
)

var (
	client                       *restclient.Client
	logWriter                    io.WriteCloser
	errFailedToConnectToGlusterd = `Failed to connect to glusterd. Please check if
- Glusterd is running(%s://%s:%d) and reachable from this node.
- Make sure hostname/IP and Port specified in the command are valid
`
)

func initRESTClient(hostname string, cacert string, insecure bool) {
	client = restclient.New(hostname, "", "", cacert, insecure)
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

func handleGlusterdConnectFailure(msg string, err error, https bool, host string, port int, errcode int) {
	if isConnectionRefusedErr(err) || isNoSuchHostErr(err) || isNoRouteToHostErr(err) {
		scheme := "http"
		if https {
			scheme = "https"
		}
		os.Stderr.WriteString(msg + "\n\n")
		os.Stderr.WriteString(fmt.Sprintf(errFailedToConnectToGlusterd, scheme, flagHostname, flagPort))
		os.Exit(errcode)
	}
}

func failure(msg string, err error, errcode int) {
	handleGlusterdConnectFailure(msg, err, flagHTTPS, flagHostname, flagPort, errcode)

	// If different error
	os.Stderr.WriteString(msg + "\n")
	if err != nil {
		os.Stderr.WriteString("\nError: " + err.Error() + "\n")
	}
	os.Exit(errcode)
}
