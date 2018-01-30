package rebalance

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"
	log "github.com/sirupsen/logrus"
)

const (
	rebalancePrefix string = "rebalance/"
)

// TimeVal represents the time value
type TimeVal struct {
	TVsec  uint64
	TVusec uint64
}

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

	tv := new(TimeVal)
	t := time.Now()
	tv.TVsec = uint64(t.Day() * t.Hour() * t.Minute() * t.Second())
	tv.TVusec = uint64((tv.TVsec * 1000000))
	hash = tv.TVsec << 3

	/*
	   Make sure at least one of those low-order bits is set.  The extra
	   shifting is because not all machines have sub-millisecond time
	   resolution.
	*/

	hash |= 1 << ((tv.TVusec >> 10) % 3)
	return hash
}

// GetRebalanceInfo provides stored rebalance details
func GetRebalanceInfo(Volname string) (*rebalanceapi.RebalInfo, error) {
	var r rebalanceapi.RebalInfo
	resp, e := store.Store.Get(context.TODO(), rebalancePrefix+Volname)
	if e != nil {
		log.WithError(e).Error("Couldn't retrieve rebalance info from store")
		return nil, e
	}

	if resp.Count != 1 {
		log.WithField("volume", Volname).Error("Rebalance info not found for the volume or rebalance process is not started for this volume")
		return nil, errors.New("Rebalance info not found for the volume or rebalance process is not started for this volume")
	}

	if e = json.Unmarshal(resp.Kvs[0].Value, &r); e != nil {
		log.WithError(e).Error("Failed to unmarshal the data into rebalance info object")
		return nil, e
	}
	return &r, nil
}

func storeRebalanceDetails(c transaction.TxnCtx) error {

	var rinfo rebalanceapi.RebalInfo
	if err := c.Get("rinfo", &rinfo); err != nil {
		return err
	}
	json, e := json.Marshal(&rinfo)
	if e != nil {
		log.WithField("error", e).Error("Failed to marshal the rebalance info object")
		return e
	}

	_, e = store.Store.Put(context.TODO(), rebalancePrefix+rinfo.Volname, string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add rebalance info to store")
		return e
	}

	return nil
}

func checkCmd(req *rebalanceapi.StartReq) rebalanceapi.Command {
	var rebalance rebalanceapi.RebalInfo
	if req.Fixlayout == true {
		rebalance.Cmd = rebalanceapi.CmdFixlayoutStart
	} else if req.Force == true {
		rebalance.Cmd = rebalanceapi.CmdStartForce
	} else {
		rebalance.Cmd = rebalanceapi.CmdStart
	}
	return rebalance.Cmd
}
