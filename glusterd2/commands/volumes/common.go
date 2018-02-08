package volumecommands

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	volgen "github.com/gluster/glusterd2/glusterd2/volgen2"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/glusterd2/xlator/options"
	"github.com/gluster/glusterd2/pkg/api"
)

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

func storeVolume(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := volume.AddOrUpdateVolumeFunc(&volinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("storeVolume: failed to store volume info")
		return err
	}

	if err := volgen.Generate(); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("generateVolfiles: failed to generate volfiles")
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
