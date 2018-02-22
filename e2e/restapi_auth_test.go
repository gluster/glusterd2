package e2e

import (
	"io/ioutil"
	"testing"

	"github.com/gluster/glusterd2/pkg/restclient"

	"github.com/stretchr/testify/require"
)

func TestRESTAPIAuth(t *testing.T) {
	r := require.New(t)

	g1, err := spawnGlusterd("./config/4.toml", true)
	r.Nil(err)
	defer g1.Stop()
	defer g1.EraseWorkdir()
	r.True(g1.IsRunning())

	secret, err := ioutil.ReadFile(g1.Workdir + "/auth")
	r.Nil(err)

	client := restclient.New("http://"+g1.ClientAddress, "glustercli", string(secret), "", false)

	// Get Peers information, should work if auth works
	peers, err := client.Peers()
	r.Nil(err)
	r.Len(peers, 1)
}
