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
	var err error
	client, err = restclient.New(hostname, user, secret, cacert, insecure)
	if err != nil {
		failure("failed to setup client", err, 1)
	}
	client.SetTimeout(time.Duration(GlobalFlag.Timeout) * time.Second)
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

	handleGlusterdConnectFailure(msg, GlobalFlag.Endpoints[0], err, errcode)

	w := os.Stderr

	w.WriteString(msg + "\n")

	if client == nil && err != nil {
		fmt.Fprintln(w, err)
		os.Exit(errcode)
	}

	resp := client.LastErrorResponse()

	if resp == nil && err != nil {
		fmt.Fprintln(w, err)
		os.Exit(errcode)
	}

	if err != nil {
		w.WriteString("\nResponse headers:\n")
		for k, v := range resp.Header {
			if strings.HasSuffix(k, "-Id") {
				w.WriteString(fmt.Sprintf("%s: %s\n", k, v[0]))
			}
		}

		w.WriteString("\nResponse body:\n")
		w.WriteString(fmt.Sprintf("%s\n", err.Error()))
	}

	os.Exit(errcode)
}
