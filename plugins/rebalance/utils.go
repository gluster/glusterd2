package rebalance

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gluster/glusterd2/glusterd2/store"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"

	log "github.com/sirupsen/logrus"
)

const (
	rebalancePrefix string = "rebalance/"
)

var (
	hash uint64
)

func setCommitHash() uint64 {

	/*
	   We need a commit hash that won't conflict with others we might have
	   set, or zero which is the implicit value if we never have.  Using
	   seconds<<3 like this ensures that we'll only get a collision if two
	   consecutive rebalances are separated by exactly 2^29 seconds - about
	   17 years - and even then there's only a 1/8 chance of a collision in
	   the low order bits.  It's far more likely that this code will have
	   changed completely by then.  If not, call me in 2031.
	   P.S. Time zone changes?  Yeah, right.
	*/

	var tsec uint64
	var tnsec uint64

	t := time.Now()
	tsec = uint64(t.Unix())
	tnsec = uint64(t.UnixNano())

	hash = tsec << 3

	/*
	   Make sure at least one of those low-order bits is set.  The extra
	   shifting is because not all machines have sub-millisecond time
	   resolution.
	*/

	hash |= 1 << ((tnsec >> 10) % 3)
	return hash
}

// GetRebalanceInfo gets the stored rebalance details
func GetRebalanceInfo(volname string) (*rebalanceapi.RebalInfo, error) {

	var rebalinfo rebalanceapi.RebalInfo

	resp, err := store.Get(context.TODO(), rebalancePrefix+volname)
	if err != nil {
		log.WithError(err).Error("Couldn't retrieve rebalance info from store")
		return nil, err
	}

	if resp.Count != 1 {
		log.WithField("volume", volname).Error("Rebalance info not found for the volume or rebalance process is not started for this volume")
		return nil, errors.New("rebalance info not found for the volume or rebalance process is not started for this volume")
	}

	if err = json.Unmarshal(resp.Kvs[0].Value, &rebalinfo); err != nil {
		log.WithError(err).Error("Failed to unmarshal the data into rebalance info object")
		return nil, err
	}
	return &rebalinfo, nil
}

// StoreRebalanceInfo : Stores the rebal info
func StoreRebalanceInfo(rinfo *rebalanceapi.RebalInfo) error {
	json, err := json.Marshal(&rinfo)
	if err != nil {
		log.WithError(err).Error("Failed to marshal the rebalance info object")
		return err
	}

	_, err = store.Put(context.TODO(), rebalancePrefix+rinfo.Volname, string(json))
	if err != nil {
		log.WithError(err).Error("Couldn't add rebalance info to store")
		return err
	}
	return nil
}

func getCmd(req *rebalanceapi.StartReq) rebalanceapi.Command {

	switch req.Option {
	case "fix-layout":
		return rebalanceapi.CmdFixLayoutStart
	case "force":
		return rebalanceapi.CmdStartForce
	case "":
		return rebalanceapi.CmdStart
	default:
		return rebalanceapi.CmdNone
	}
}
