// Package gdctx is the runtime context of GlusterD
//
// This file implements the global runtime context for GlusterD.
// Any package that needs access to the GlusterD global runtime context just
// needs to import this package.
package gdctx

import (
	"crypto/rand"
	"expvar"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strconv"

	"github.com/gluster/glusterd2/pkg/utils"
	"github.com/gluster/glusterd2/version"

	config "github.com/spf13/viper"
)

var (
	expOpVersion = expvar.NewInt("op_version")
)

// Any object that is a part of the GlusterD context and needs to be available
// to other packages should be declared here as exported global variables
var (
	OpVersion          int
	HostIP             string
	HostName           string
	LocalAuthToken     string
	RESTAPIAuthEnabled = false
)

// SetHostnameAndIP will initialize HostIP and HostName global variables
func SetHostnameAndIP() error {
	hostIP, err := utils.GetLocalIP()
	if err != nil {
		return err
	}
	HostIP = hostIP

	hostName, err := os.Hostname()
	if err != nil {
		return err
	}
	HostName = hostName

	return nil
}

func init() {
	OpVersion = version.MaxOpVersion
	expOpVersion.Set(int64(OpVersion))
}

// GenerateLocalAuthToken generates random secret if not already generated
func GenerateLocalAuthToken() error {
	if config.IsSet("restauth") && !config.GetBool("restauth") {
		return nil
	}

	RESTAPIAuthEnabled = true
	authFile := path.Join(config.GetString("localstatedir"), "auth")

	_, err := os.Stat(authFile)
	if os.IsNotExist(err) {
		data := make([]byte, 32)
		_, err := rand.Read(data)
		if err == nil {
			LocalAuthToken = fmt.Sprintf("%x", data)
			if errWrite := ioutil.WriteFile(authFile, []byte(LocalAuthToken), 0640); errWrite != nil {
				return errWrite
			}
			err := protectAuthFile(authFile)
			if err != nil {
				return err
			}
		}
		return err
	} else if err == nil {
		secret, err := ioutil.ReadFile(authFile)
		if err != nil {
			return err
		}
		LocalAuthToken = string(secret)
	} else {
		return err
	}
	return nil
}

func protectAuthFile(authfile string) error {
	var uGID string

	cuser, err := user.Current()
	if err != nil {
		return err
	}

	uGID = cuser.Gid

	usr, err := user.LookupGroup("gluster")
	if err != nil {
		if _, ok := err.(user.UnknownGroupError); !ok {
			return err
		}
	} else {
		uGID = usr.Gid
	}

	gID, err := strconv.Atoi(uGID)
	if err != nil {
		return err
	}

	uID, err := strconv.Atoi(cuser.Uid)
	if err != nil {
		return err
	}
	err = os.Chown(authfile, uID, gID)
	return err
}
