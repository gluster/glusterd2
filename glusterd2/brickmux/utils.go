package brickmux

import (
	"fmt"
	config "github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// getBrickVolfilePath returns path correspponding to the volfileID. Since, brick volfiles are now stored on
// disk, added a check to ensure the volfile exists on the path.
func getBrickVolfilePath(volfileID string) (string, error) {

	volfilePath := path.Join(config.GetString("localstatedir"),
		"volfiles", volfileID+".vol")

	if _, err := os.Stat(volfilePath); os.IsNotExist(err) {
		return "", err
	}

	return volfilePath, nil
}

// glusterdGetSockFromBrickPid returns the socket file from the /proc/
// filesystem for the concerned process running with the same pid
func glusterdGetSockFromBrickPid(pid int) (string, error) {
	content, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", err
	}

	parts := strings.Split(string(content), "\x00")
	prevPart := ""
	socketFile := ""
	for _, p := range parts {
		if prevPart == "-S" {
			socketFile = p
			break
		}
		prevPart = p
	}
	return socketFile, nil
}
