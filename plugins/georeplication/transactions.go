package georeplication

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"

	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

const (
	gsyncdStatusTxnKey string = "gsyncdstatuses"
)

func txnGeorepCreate(c transaction.TxnCtx) error {
	var sessioninfo georepapi.GeorepSession
	if err := c.Get("geosession", &sessioninfo); err != nil {
		return err
	}

	if err := addOrUpdateSession(&sessioninfo); err != nil {
		c.Logger().WithError(err).WithField(
			"mastervolid", sessioninfo.MasterID).WithField(
			"remotevolid", sessioninfo.RemoteID).Debug(
			"failed to store Geo-replication info")
		return err
	}

	return nil
}

func gsyncdAction(c transaction.TxnCtx, action actionType) error {
	var masterid string
	var remoteid string
	if err := c.Get("mastervolid", &masterid); err != nil {
		return err
	}
	if err := c.Get("remotevolid", &remoteid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, remoteid)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{
		"master": sessioninfo.MasterVol,
		"remote": sessioninfo.RemoteHosts[0].Hostname + "::" + sessioninfo.RemoteVol,
	}).Info(action.String() + " gsyncd monitor")

	gsyncdDaemon, err := newGsyncd(*sessioninfo)
	if err != nil {
		return err
	}

	switch action {
	case actionStart:
		// Create Geo-replication Log dir if not exists
		err = os.MkdirAll(path.Join(config.GetString("logdir"), "glusterfs", "geo-replication"), os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
		err = configFileGenerate(sessioninfo)
		if err != nil {
			return err
		}
		err = daemon.Start(gsyncdDaemon, true, c.Logger())
	case actionStop:
		err = daemon.Stop(gsyncdDaemon, true, c.Logger())
	case actionPause:
		err = daemon.Signal(gsyncdDaemon, syscall.SIGSTOP, c.Logger())
	case actionResume:
		err = daemon.Signal(gsyncdDaemon, syscall.SIGCONT, c.Logger())
	}

	return err
}

func txnGeorepStart(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionStart)
}

func txnGeorepStop(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionStop)
}

func txnGeorepDelete(c transaction.TxnCtx) error {
	var masterid string
	var remoteid string
	if err := c.Get("mastervolid", &masterid); err != nil {
		return err
	}
	if err := c.Get("remotevolid", &remoteid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, remoteid)
	if err != nil {
		return err
	}

	if err := deleteSession(masterid, remoteid); err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"master": sessioninfo.MasterVol,
			"remote": sessioninfo.RemoteHosts[0].Hostname + "::" + sessioninfo.RemoteVol,
		}).Debug("failed to delete Geo-replication info from store")
		return err
	}

	return nil
}

func txnGeorepPause(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionPause)
}

func txnGeorepResume(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionResume)
}

func txnGeorepStatus(c transaction.TxnCtx) error {
	var masterid string
	var remoteid string
	var err error

	if err = c.Get("mastervolid", &masterid); err != nil {
		return err
	}

	if err = c.Get("remotevolid", &remoteid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, remoteid)
	if err != nil {
		return err
	}

	// Get Master vol info to get the bricks List
	volinfo, err := volume.GetVolume(sessioninfo.MasterVol)
	if err != nil {
		return err
	}

	var workersStatuses = make(map[string]georepapi.GeorepWorker)

	for _, w := range volinfo.GetLocalBricks() {
		gsyncd, err := newGsyncd(*sessioninfo)
		if err != nil {
			return err
		}
		args := gsyncd.statusArgs(w.Path)

		out, err := utils.ExecuteCommandOutput(gsyncdCommand, args...)
		if err != nil {
			return err
		}

		var worker georepapi.GeorepWorker
		if err = json.Unmarshal(out, &worker); err != nil {
			return err
		}

		// Unique key for master brick UUID:BRICK_PATH
		key := gdctx.MyUUID.String() + ":" + w.Path
		workersStatuses[key] = worker
	}

	c.SetNodeResult(gdctx.MyUUID, gsyncdStatusTxnKey, workersStatuses)
	return nil
}

func aggregateGsyncdStatus(ctx transaction.TxnCtx, nodes []uuid.UUID) (*map[string]georepapi.GeorepWorker, error) {
	var workersStatuses = make(map[string]georepapi.GeorepWorker)

	// Loop over each node on which txn was run.
	// Fetch brick statuses stored by each node in transaction context.
	for _, node := range nodes {
		var tmp = make(map[string]georepapi.GeorepWorker)
		err := ctx.GetNodeResult(node, gsyncdStatusTxnKey, &tmp)
		if err != nil {
			return nil, errors.New("aggregateGsyncdStatus: Could not fetch results from transaction context")
		}

		// Single final Hashmap
		for k, v := range tmp {
			workersStatuses[k] = v
		}
	}

	return &workersStatuses, nil
}

func txnGeorepConfigSet(c transaction.TxnCtx) error {
	var masterid string
	var remoteid string
	var session georepapi.GeorepSession

	if err := c.Get("mastervolid", &masterid); err != nil {
		return err
	}
	if err := c.Get("remotevolid", &remoteid); err != nil {
		return err
	}

	if err := c.Get("session", &session); err != nil {
		return err
	}

	if err := addOrUpdateSession(&session); err != nil {
		c.Logger().WithError(err).WithField(
			"mastervolid", session.MasterID).WithField(
			"remotevolid", session.RemoteID).Debug(
			"failed to store Geo-replication info")
		return err
	}

	return nil
}

func configFileGenerate(session *georepapi.GeorepSession) error {
	confdata := []string{"[vars]"}
	var err error

	gsyncdDaemon, err := newGsyncd(*session)
	if err != nil {
		return err
	}
	configFile := gsyncdDaemon.ConfigFile()

	// Create Config dir if not exists
	err = os.MkdirAll(path.Dir(configFile), os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	vol, err := volume.GetVolume(session.MasterVol)
	if err != nil {
		return err
	}

	// Remote host and UUID details
	var remote = make([]string, 0, len(session.RemoteHosts))
	for _, sh := range session.RemoteHosts {
		remote = append(remote, sh.PeerID.String()+":"+sh.Hostname)
	}
	confdata = append(confdata,
		fmt.Sprintf("slave-bricks=%s", strings.Join(remote, ",")),
	)

	// Master Bricks details
	bricks := vol.GetBricks()
	var master = make([]string, 0, len(bricks))
	for _, b := range bricks {
		master = append(master, b.PeerID.String()+":"+b.Hostname+":"+b.Path)
	}
	confdata = append(confdata,
		fmt.Sprintf("master-bricks=%s", strings.Join(master, ",")),
	)

	// Master Volume ID
	confdata = append(confdata,
		fmt.Sprintf("master-volume-id=%s", session.MasterID.String()),
	)

	// Remote Volume ID
	confdata = append(confdata,
		fmt.Sprintf("slave-volume-id=%s", session.RemoteID.String()),
	)

	// Master Replica Count
	confdata = append(confdata,
		fmt.Sprintf("master-replica-count=%d", vol.Subvols[0].ReplicaCount),
	)

	confdata = append(confdata,
		fmt.Sprintf("master-disperse-count=%d", vol.Subvols[0].DisperseCount),
	)

	// Custom session configurations if any
	for k, v := range session.Options {
		confdata = append(confdata, k+"="+v)
	}

	return ioutil.WriteFile(configFile, []byte(strings.Join(confdata, "\n")), 0644)
}

func txnGeorepConfigFilegen(c transaction.TxnCtx) error {
	var masterid string
	var remoteid string
	var session georepapi.GeorepSession
	var restartRequired bool
	var err error

	if err = c.Get("mastervolid", &masterid); err != nil {
		return err
	}
	if err = c.Get("remotevolid", &remoteid); err != nil {
		return err
	}

	if err = c.Get("session", &session); err != nil {
		return err
	}

	if err = c.Get("restartRequired", &restartRequired); err != nil {
		return err
	}

	if restartRequired {

		if err = gsyncdAction(c, actionStop); err != nil {
			return err
		}

		if err = gsyncdAction(c, actionStart); err != nil {
			return err
		}
	} else {
		// Restart not required, Generate config file Gsynd will reload
		// automatically if running
		if err = configFileGenerate(&session); err != nil {
			return err
		}
	}
	return nil
}

func txnSSHKeysGenerate(c transaction.TxnCtx) error {
	var volname string
	var err error
	var args []string

	if err = c.Get("volname", &volname); err != nil {
		return err
	}

	secretPemFile := path.Join(
		config.GetString("localstatedir"),
		"geo-replication",
		"secret.pem",
	)
	tarSSHPemFile := path.Join(
		config.GetString("localstatedir"),
		"geo-replication",
		"tar_ssh.pem",
	)

	// Create Directory if not exists

	if err = os.MkdirAll(path.Dir(secretPemFile), os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	sshkey := georepapi.GeorepSSHPublicKey{PeerID: gdctx.MyUUID}

	// Generate secret.pem file if not available
	if _, err := os.Stat(secretPemFile); os.IsNotExist(err) {
		args = []string{"-N", "", "-f", secretPemFile}
		_, err = utils.ExecuteCommandOutput("ssh-keygen", args...)
		if err != nil {
			return err
		}
	}

	data, err := ioutil.ReadFile(secretPemFile + ".pub")
	if err != nil {
		return err
	}
	sshkey.GsyncdKey = string(data)

	// Generate tar_ssh.pem file if not available
	if _, err := os.Stat(tarSSHPemFile); os.IsNotExist(err) {
		args = []string{"-N", "", "-f", tarSSHPemFile}
		_, err = utils.ExecuteCommandOutput("ssh-keygen", args...)
		if err != nil {
			return err
		}
	}
	if data, err = ioutil.ReadFile(tarSSHPemFile + ".pub"); err != nil {
		return err
	}
	sshkey.TarKey = string(data)

	err = addOrUpdateSSHKey(volname, sshkey)

	return err
}

func txnSSHKeysPush(c transaction.TxnCtx) error {
	var err error
	var sshkeys []georepapi.GeorepSSHPublicKey
	var user string

	if err = c.Get("sshkeys", &sshkeys); err != nil {
		return err
	}

	if err = c.Get("user", &user); err != nil {
		return err
	}

	sshCmdGsyncdPrefix := "command=\"" + gsyncdCommand + "\"  "
	sshCmdTarPrefix := "command=\"tar ${SSH_ORIGINAL_COMMAND#* }\"  "
	authorizedKeysFile := "/root/.ssh/authorized_keys"

	if user != "root" {
		authorizedKeysFile = "/home/" + user + "/.ssh/authorized_keys"
	}

	// Prepare Public Keys(Prefix GSYNCD_CMD to Gsyncd key and Tar CMD)
	keysToAdd := make([]string, len(sshkeys)*2)
	keynum := 0
	for _, key := range sshkeys {
		keysToAdd[keynum] = sshCmdGsyncdPrefix + key.GsyncdKey
		keynum++
		keysToAdd[keynum] = sshCmdTarPrefix + key.TarKey
		keynum++
	}

	// TODO: Handle if authorized_keys is configured to different location
	// TODO: Set permissions and SELinux contexts

	contentRaw, err := ioutil.ReadFile(authorizedKeysFile)
	if err != nil {
		return err
	}
	content := string(contentRaw)

	// Append if not exists in authorized_keys file
	for _, key := range keysToAdd {
		if !strings.Contains(content, key) {
			content = content + key
		}
	}

	err = ioutil.WriteFile(authorizedKeysFile+".tmp", []byte(content), 600)
	if err != nil {
		return err
	}

	err = os.Rename(authorizedKeysFile+".tmp", authorizedKeysFile)
	if err != nil {
		return err
	}

	return nil
}
