package quota

import (
	"fmt"

	daemon "github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/errors"
)

const (
	quotaEnabledKey = "quota.enable"
)

func validateOptions(v *volume.Volinfo, key, value string) error {

	if v.State != volume.VolStarted {
		return errors.ErrVolNotStarted
	}

	switch key {
	case "enable":
		switch value {
		case "on":
			// As quotad for various process share a
			// single pid, it is necessary to check the
			// options too.
			if isQuotadRunning() && isQuotaEnabled(v) {
				return errors.ErrProcessAlreadyRunning
			}
		case "off":
			// As quotad for various process share a
			// single pid, it is necessary to check the
			// options too.
			if !isQuotaEnabled(v) {
				return errors.ErrQuotadNotEnabled
			}
			if !isQuotadRunning() {
				return errors.ErrQuotadNotRunning
			}
		}
		return nil
	case "deem-statfs":
		fallthrough
	case "hard-timeout":
		fallthrough
	case "soft-timeout":
		fallthrough
	case "alert-time":
		fallthrough
	case "default-soft-limit":
		if !isQuotaEnabled(v) {
			return fmt.Errorf("quota must be enabled to set '%s' option", key)
		}
	default:
		return fmt.Errorf("'%s' is not a valid quota option", key)
	}

	return nil
}

// isQuotaEnabled is used to check if the quota option is enabled for
// that particular volume.
func isQuotaEnabled(v *volume.Volinfo) bool {
	val, exists := v.Options[quotaEnabledKey]
	if exists && val == "on" {
		return true
	}
	return false
}

// isQuotadRunning returns true for running and false for
// failures and not running on that machine.
// It is not volume based.
func isQuotadRunning() bool {
	quotadDaemon, err := NewQuotad()
	if err != nil {
		return false
	}
	pid, err := daemon.ReadPidFromFile(quotadDaemon.PidFile())
	if err != nil {
		return false
	}

	if _, err = daemon.GetProcess(pid); err != nil {
		return false
	}
	return true
}

func init() {
	xlator.RegisterValidationFunc(name, validateOptions)
}
