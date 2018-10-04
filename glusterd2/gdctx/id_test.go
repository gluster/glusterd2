package gdctx

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	invalidUUID = "this-is-an-invalid-uuid"
)

var (
	testStateDir string
)

func resetEnv(t *testing.T) {
	MyUUID = nil
	MyClusterID = nil
	require.NoError(t, os.Setenv(envClusterIDKey, ""))
	require.NoError(t, os.Setenv(envPeerIDKey, ""))

	require.NoError(t, os.RemoveAll(testStateDir))
	require.NoError(t, os.MkdirAll(testStateDir, os.ModePerm))
	config.Set("localstatedir", testStateDir)

}

// TestIDs tests various ways the peer and cluster ids can be loaded, updated and saved
func TestIDs(t *testing.T) {
	d, err := ioutil.TempDir("", "TestGdtxUUID")
	require.NoError(t, err)
	defer os.RemoveAll(d)

	testStateDir = d

	t.Run("Init", testInitUUID)
	t.Run("SaveToFile", testSaveFile)
	t.Run("ReloadFromFile", testReloadFile)
	t.Run("LoadFromENV", testEnvIDs)
	t.Run("UpdateClusterID", testUpdateClusterID)
}

// testInitUUID ensures that InitUUID properly sets MyUUID and MyClusterID
func testInitUUID(t *testing.T) {
	resetEnv(t)

	// Empty run
	require.NoError(t, InitUUID())

	// Ensure that both uuids are not nil anymore
	require.NotNil(t, MyUUID)
	require.False(t, uuid.Equal(uuid.NIL, MyUUID))
	require.NotNil(t, MyUUID)
	require.False(t, uuid.Equal(uuid.NIL, MyClusterID))
}

// testUpdateClusterID ensures that UpdateClusterID properly updates and saves
// the cluster-id to the given UUID
func testUpdateClusterID(t *testing.T) {
	resetEnv(t)

	clusterID := uuid.NewRandom()
	clusterIDstr := clusterID.String()

	// Ensure that a valid uuid is set and updated
	require.NoError(t, UpdateClusterID(clusterIDstr))
	require.True(t, uuid.Equal(MyClusterID, clusterID))

	// Ensure that an invalid uuid cannot be set
	require.Error(t, UpdateClusterID("this-is-an-invalid-uuid"))
}

// TestSaveFile tests saving and reloading uuid from file
func testSaveFile(t *testing.T) {
	resetEnv(t)

	// Storing the randomly initialized ids in a fresh uuidConfig and saving it file
	c1 := newUUIDConfig()
	peerID := c1.GetString(peerIDKey)
	clusterID := c1.GetString(clusterIDKey)

	require.NoError(t, c1.save())

	// Create a new uuidConfig that will load values from the saved file
	c2 := newUUIDConfig()
	require.NoError(t, c2.reload(false))

	require.Equal(t, peerID, c2.GetString(peerIDKey))
	require.Equal(t, clusterID, c2.GetString(clusterIDKey))
}

// testReloadFile tests if reloading uuid file works correctly in different cases
// NOTE: Successful case is being tested in testSaveFile
func testReloadFile(t *testing.T) {
	t.Run("NotPresent", testReloadFileNoFile)
	t.Run("Empty", testReloadFileEmptyFile)
	t.Run("InvalidTOML", testReloadFileInvalidTOML)
	t.Run("InvalidUUID", testReloadFileInvalidUUID)
}

func testReloadFileNoFile(t *testing.T) {
	resetEnv(t)

	c1 := newUUIDConfig()

	// Reloading the file should fail if it is missing, except during initialization
	require.Error(t, c1.reload(false))
	require.NoError(t, c1.reload(true))
}

func testReloadFileEmptyFile(t *testing.T) {
	resetEnv(t)

	f, err := os.Create(uuidFilePath())
	require.NoError(t, err)
	require.NoError(t, f.Close())

	c1 := newUUIDConfig()
	// Reloading empty file should always succeed
	require.NoError(t, c1.reload(false))
	require.NoError(t, c1.reload(true))
}

func testReloadFileInvalidTOML(t *testing.T) {
	resetEnv(t)

	f, err := os.Create(uuidFilePath())
	require.NoError(t, err)
	_, err = f.WriteString("this: is: not: toml\nno really")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	c1 := newUUIDConfig()
	// Reloading file with invalid toml should always fail
	require.Error(t, c1.reload(false))
	require.Error(t, c1.reload(true))
}

func testReloadFileInvalidUUID(t *testing.T) {
	resetEnv(t)

	c1 := newUUIDConfig()
	// Set invalid uuid as config and save to file
	c1.Set(peerIDKey, invalidUUID)
	c1.Set(clusterIDKey, invalidUUID)
	require.NoError(t, c1.save())

	c2 := newUUIDConfig()
	// Reloading file with invalid uuids should fail always
	require.Error(t, c2.reload(false))
	require.Error(t, c2.reload(true))
}

// testEnvIDs tests getting the ids from the environment
func testEnvIDs(t *testing.T) {
	t.Run("ValidUUIDs", testEnvIDsValid)
	t.Run("InvalidUUIDs", testEnvIDsInvalid)
}

func testEnvIDsValid(t *testing.T) {
	resetEnv(t)

	// Ensure that valid ids are loaded from the environment
	peerID := uuid.NewRandom()
	clusterID := uuid.NewRandom()

	require.NoError(t, os.Setenv(envPeerIDKey, peerID.String()))
	require.NoError(t, os.Setenv(envClusterIDKey, clusterID.String()))

	require.NoError(t, InitUUID())
	require.True(t, uuid.Equal(peerID, MyUUID))
	require.True(t, uuid.Equal(clusterID, MyClusterID))
}

func testEnvIDsInvalid(t *testing.T) {
	// Ensure that invalid ids are not loaded
	resetEnv(t)
	require.NoError(t, os.Setenv(envPeerIDKey, invalidUUID))
	require.Error(t, InitUUID())

	resetEnv(t)
	require.NoError(t, os.Setenv(envClusterIDKey, invalidUUID))
	require.Error(t, InitUUID())
}
