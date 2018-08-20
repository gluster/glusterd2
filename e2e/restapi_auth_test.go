package e2e

import (
	"io/ioutil"
	"testing"

	"github.com/gluster/glusterd2/pkg/restclient"

	"github.com/stretchr/testify/require"
)

func TestRESTAPIAuth(t *testing.T) {
	r := require.New(t)

	tc, err := setupCluster("./config/4.toml")
	r.NoError(err)
	defer teardownCluster(tc)

	g1 := tc.gds[0]
	r.True(g1.IsRunning())

	secret, err := ioutil.ReadFile(g1.LocalStateDir + "/auth")
	r.Nil(err)

	client, err := restclient.New("http://"+g1.ClientAddress, "glustercli", string(secret), "", false)
	r.Nil(err)
	r.NotNil(client)

	// Get Peers information, should work if auth works
	peers, err := client.Peers()
	r.Nil(err)
	r.Len(peers, 1)
}
