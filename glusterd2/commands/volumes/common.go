package volumecommands

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	volgen "github.com/gluster/glusterd2/glusterd2/volgen2"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const volumeIDXattrKey = "trusted.glusterfs.volume-id"

// validateOptions validates if the options and their values are valid and can
// be set on a volume.
func validateOptions(opts map[string]string) error {

	for k, v := range opts {
		o, err := xlator.FindOption(k)
		if err != nil {
			return err
		}

		if err := o.Validate(v); err != nil {
			return err
		}
		// TODO: Check op-version
	}

	return nil
}

func validateXlatorOptions(opts map[string]string, volinfo *volume.Volinfo) error {
	for k, v := range opts {
		_, xl, key, err := options.SplitKey(k)
		if err != nil {
			return err
		}
		xltr, err := xlator.Find(xl)
		if err != nil {
			return err
		}
		if xltr.Validate != nil {
			if err := xltr.Validate(volinfo, key, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func expandOptions(opts map[string]string) (map[string]string, error) {
	resp, err := store.Store.Get(context.TODO(), "groupoptions")
	if err != nil {
		return nil, err
	}

	var groupOptions map[string][]api.VolumeOption
	if err := json.Unmarshal(resp.Kvs[0].Value, &groupOptions); err != nil {
		return nil, err
	}

	options := make(map[string]string)
	for opt, val := range opts {
		optionSet, ok := groupOptions[opt]
		if !ok {
			options[opt] = val
		} else {
			for _, option := range optionSet {
				switch val {
				case "on":
					options[option.Name] = option.OnValue
				case "off":
					options[option.Name] = option.OffValue
				default:
					return nil, errors.New("Need either on or off")
				}
			}
		}
	}
	return options, nil
}

func notifyVolfileChange(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if volinfo.State != volume.VolStarted {
		return nil
	}

	sunrpc.FetchSpecNotify(c)

	return nil
}

func validateBricks(c transaction.TxnCtx) error {

	var err error

	var bricks []brick.Brickinfo
	if err = c.Get("bricks", &bricks); err != nil {
		return err
	}

	var checks brick.InitChecks
	if err = c.Get("brick-checks", &checks); err != nil {
		return err
	}

	for _, b := range bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		if err = b.Validate(checks); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("Brick validation failed")
			return err
		}
	}

	return nil
}

func initBricks(c transaction.TxnCtx) error {

	var err error

	var bricks []brick.Brickinfo
	if err = c.Get("bricks", &bricks); err != nil {
		return err
	}

	var checks brick.InitChecks
	if err = c.Get("brick-checks", &checks); err != nil {
		return err
	}

	flags := 0
	if checks.IsInUse {
		// Perform a pure replace operation, which fails if the named
		// attribute does not already exist.
		flags = unix.XATTR_CREATE
	}
	for _, b := range bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		err = unix.Setxattr(b.Path, volumeIDXattrKey, []byte(b.VolumeID), flags)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"path": b.Path,
				"key":  volumeIDXattrKey}).Error("Setxattr failed")
			return err
		}

		path := filepath.Join(b.Path, ".glusterfs")
		err = os.MkdirAll(path, os.ModeDir|os.ModePerm)
		if err != nil {
			log.WithError(err).WithField(
				"path", path).Error("MkdirAll failed")
			return err
		}
	}

	return nil
}

func undoInitBricks(c transaction.TxnCtx) error {

	var bricks []brick.Brickinfo
	if err := c.Get("bricks", &bricks); err != nil {
		return err
	}

	// FIXME: This is prone to races. See issue #314

	for _, b := range bricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		unix.Removexattr(b.Path, volumeIDXattrKey)
		os.Remove(filepath.Join(b.Path, ".glusterfs"))
	}

	return nil
}

func storeVolume(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volinfo").Debug("Failed to get key from store")
		return err
	}

	if err := volume.AddOrUpdateVolumeFunc(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("failed to store volume info")
		return err
	}

	if err := volgen.Generate(); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("failed to generate volfiles")
		return err
	}

	return nil
}

// LoadDefaultGroupOptions loads the default group option map into the store
func LoadDefaultGroupOptions() error {
	groupOptions, err := json.Marshal(defaultGroupOptions)
	if err != nil {
		return err
	}
	if _, err := store.Store.Put(context.TODO(), "groupoptions", string(groupOptions)); err != nil {
		return err
	}
	return nil
}
