package volumecommands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const volumeIDXattrKey = "trusted.glusterfs.volume-id"

// validateOptions validates if the options and their values are valid and can
// be set on a volume.
func validateOptions(opts map[string]string, adv, exp, dep bool) error {

	for k, v := range opts {
		o, err := xlator.FindOption(k)
		if err != nil {
			return err
		}

		switch {
		case !o.IsSettable():
			return fmt.Errorf("Option %s cannot be set", k)

		case o.IsAdvanced() && !adv:
			return fmt.Errorf("Option %s is an advanced option. To set it pass the advanced flag", k)

		case o.IsExperimental() && !exp:
			return fmt.Errorf("Option %s is an experimental option. To set it pass the experimental flag", k)

		case o.IsDeprecated() && !dep:
			// TODO: Return deprecation version and alternative option if available
			return fmt.Errorf("Option %s will be deprecated in future releases. To set it pass the deprecated flag", k)
		}

		if err := o.Validate(v); err != nil {
			return fmt.Errorf("Failed to validate value(%s) for key(%s): %s", k, v, err.Error())
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

func expandGroupOptions(opts map[string]string) (map[string]string, error) {
	resp, err := store.Get(context.TODO(), "groupoptions")
	if err != nil {
		return nil, err
	}

	var groupOptions map[string]*api.OptionGroup
	if err := json.Unmarshal(resp.Kvs[0].Value, &groupOptions); err != nil {
		return nil, err
	}

	options := make(map[string]string)
	for opt, val := range opts {
		optionSet, ok := groupOptions[opt]
		if !ok {
			options[opt] = val
		} else {
			for _, option := range optionSet.Options {
				switch val {
				case "on":
					options[option.Name] = option.OnValue
				case "off":
					op, err := xlator.FindOption(option.Name)
					if err != nil {
						return nil, err
					}
					options[option.Name] = op.DefaultValue
				default:
					return nil, errors.New("need either on or off")
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

// This txn step is used in volume create and in volume expand
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

	var allBricks []brick.Brickinfo
	if err = c.Get("all-bricks-in-cluster", &allBricks); err != nil {
		return err
	}

	var allLocalBricks []brick.Brickinfo
	for _, b := range allBricks {
		if uuid.Equal(gdctx.MyUUID, b.PeerID) {
			allLocalBricks = append(allLocalBricks, b)
		}
	}

	for _, b := range bricks {
		if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
			continue
		}

		if err = b.Validate(checks, allLocalBricks); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("Brick validation failed")
			return err
		}
	}

	return nil
}

// This txn step is used in volume create and in volume expand
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
	if checks.WasInUse {
		// Perform a pure replace operation, which fails if the named
		// attribute does not already exist.
		flags = unix.XATTR_CREATE
	}
	for _, b := range bricks {
		if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
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

// This txn step is used in volume create and in volume expand
func undoInitBricks(c transaction.TxnCtx) error {

	var bricks []brick.Brickinfo
	if err := c.Get("bricks", &bricks); err != nil {
		return err
	}

	// FIXME: This is prone to races. See issue #314

	for _, b := range bricks {
		if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
			continue
		}

		unix.Removexattr(b.Path, volumeIDXattrKey)
		os.Remove(filepath.Join(b.Path, ".glusterfs"))
	}

	return nil
}

// StoreVolume uses to store the volinfo and to generate client volfile
func storeVolume(c transaction.TxnCtx) error {
	return storeVolInfo(c, "volinfo")
}

// storeVolInfo uses to store the volinfo based on key and to generate client volfile
func storeVolInfo(c transaction.TxnCtx, key string) error {
	var volinfo volume.Volinfo
	if err := c.Get(key, &volinfo); err != nil {
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

// undoStoreVolume revert back volinfo and to generate client volfile
func undoStoreVolume(c transaction.TxnCtx) error {
	return storeVolInfo(c, "oldvolinfo")
}

// LoadDefaultGroupOptions loads the default group option map into the store
func LoadDefaultGroupOptions() error {
	groupOptions, err := json.Marshal(defaultGroupOptions)
	if err != nil {
		return err
	}
	if _, err := store.Put(context.TODO(), "groupoptions", string(groupOptions)); err != nil {
		return err
	}
	return nil
}

//validateVolumeFlags checks for Flags in volume create and expand
func validateVolumeFlags(flag map[string]bool) error {
	if len(flag) > 4 {
		return gderrors.ErrInvalidVolFlags
	}
	for key := range flag {
		switch key {
		case "reuse-bricks", "allow-root-dir", "allow-mount-as-brick", "create-brick-dir":
			continue
		default:
			return fmt.Errorf("volume flag not supported %s", key)
		}
	}
	return nil
}

func isActionStepRequired(opt map[string]string, volinfo *volume.Volinfo) bool {

	if volinfo.State != volume.VolStarted {
		return false
	}
	for k := range opt {
		_, xl, _, err := options.SplitKey(k)
		if err != nil {
			continue
		}
		if xltr, err := xlator.Find(xl); err == nil && xltr.Actor != nil {
			return true
		}
	}

	return false
}
