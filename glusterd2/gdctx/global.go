// Package gdctx is the runtime context of GlusterD
//
// This file implements the global runtime context for GlusterD.
// Any package that needs access to the GlusterD global runtime context just
// needs to import this package.
package gdctx

import (
	"expvar"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"

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
	if !config.GetBool("restauth") {
		return nil
	}

	RESTAPIAuthEnabled = true
	workdir := config.GetString("workdir")
	authFile := path.Join(workdir, "auth")
	_, err := os.Stat(authFile)
	if os.IsNotExist(err) {
		data := make([]byte, 32)
		_, err := rand.Read(data)
		if err == nil {
			LocalAuthToken = fmt.Sprintf("%x", data)
			if errWrite := ioutil.WriteFile(authFile, []byte(LocalAuthToken), 0640); errWrite != nil {
				return errWrite
			}
		}
		return err
	} else if err != nil {
		return err
	}

	return nil
}
