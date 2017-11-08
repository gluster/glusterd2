package e2e

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/gluster/glusterd2/pkg/restclient"

	"github.com/stretchr/testify/require"
)

func TestRESTAPIAuth(t *testing.T) {
	r := require.New(t)

	// FIXME: Ignored error since it makes REST call
	g1, _ := spawnGlusterd("./config/4.yaml", true)

	defer g1.Stop()
	defer g1.EraseWorkdir()
	r.True(g1.IsRunning())

	// Sleep till Glusterd spawns and generates Auth
	time.Sleep(5 * time.Second)

	secret, err := ioutil.ReadFile(g1.Workdir + "/auth")
	r.Nil(err)

	client := restclient.New("http://"+g1.ClientAddress, "glustercli", string(secret), "", false)

	// Get Peers information, should work if auth works
	peers, err := client.Peers()
	r.Nil(err)
	r.Len(peers, 1)
}
