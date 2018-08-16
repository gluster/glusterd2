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
	client                       restclient.GlusterD2Client
	logWriter                    io.WriteCloser
	errFailedToConnectToGlusterd = `Failed to connect to glusterd. Please check if
- Glusterd is running(%s) and reachable from this node.
- Make sure Endpoints specified in the command is valid
`
)

func initRESTClient(hostname, user, secret, cacert string, insecure bool) {
	gd2Client, err := restclient.New(hostname, user, secret, cacert, insecure)
	if err != nil {
		failure("failed to setup client", err, 1)
	}
	gd2Client.SetTimeout(time.Duration(GlobalFlag.Timeout) * time.Second)
	client = gd2Client
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

	w := os.Stderr
	exit := func(err error) {
		if err != nil {
			fmt.Fprintln(w, err)
		}
		os.Exit(errcode)
	}

	handleGlusterdConnectFailure(msg, GlobalFlag.Endpoints[0], err, errcode)

	fmt.Fprintln(w, msg)

	gd2client, ok := client.(*restclient.Client)
	if client == nil || !ok {
		exit(err)
	}

	resp := gd2client.LastErrorResponse()
	if resp == nil {
		exit(err)
	}

	fmt.Fprintln(w, "\nResponse headers:")
	for k, vals := range resp.Header {
		if strings.HasSuffix(k, "-Id") {
			for _, val := range vals {
				fmt.Fprintf(w, "%s: %s\n", k, val)
			}
		}
	}
	fmt.Fprintln(w, "\nResponse body:")
	exit(err)
}
