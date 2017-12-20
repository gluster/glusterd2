package rebalance

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
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

func glusterdvolinforesetstats(r RebalanceInfo) {
	r.RebalanceFiles = 0
	r.RebalanceData = 0
	r.LookedupFiles = 0
	r.RebalanceFailures = 0
	r.ElapsedTime = 0
	r.SkippedFiles = 0
	r.TimeLeft = 0
}

func setcommithash(r *RebalanceInfo) {
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
	r.CommitHash = hash
	return
}

// GetRebalanceDetails provides the details of rebalance process which is stored
func GetRebalanceDetails(rebalance *RebalanceInfo) (*RebalanceInfo, error) {
	v := new(RebalanceInfo)
	v.Volname = rebalance.Volname
	v.RebalanceID = rebalance.RebalanceID
	v.Status = rebalance.Status
	v.RebalanceFiles = rebalance.RebalanceFiles
	v.RebalanceData = rebalance.RebalanceData
	v.LookedupFiles = rebalance.LookedupFiles
	v.RebalanceFailures = rebalance.RebalanceFailures
	v.ElapsedTime = rebalance.ElapsedTime
	v.SkippedFiles = rebalance.SkippedFiles
	v.TimeLeft = rebalance.TimeLeft
	v.CommitHash = rebalance.CommitHash
	return v, nil
}

// GetRebalanceInfo fetches the json object from the store and unmarshalls it into
func GetRebalanceInfo(Volname string) (*RebalanceInfo, error) {
	var r RebalanceInfo
	resp, e := store.Store.Get(context.TODO(), rebalancePrefix+Volname)
	if e != nil {
		log.WithError(e).Error("Couldn't retrive volume from store")
		return nil, e
	}

	if resp.Count != 1 {
		log.WithField("volume", Volname).Error("volume not found")
		return nil, errors.New("volume not found")
	}

	if e = json.Unmarshal(resp.Kvs[0].Value, &r); e != nil {
		log.WithError(e).Error("Failed to unmarshal the data into rebalance info object")
		return nil, e
	}
	return &r, nil
}

func storeRebalanceDetails(c transaction.TxnCtx) error {

	var rinfo RebalanceInfo
	if err := c.Get("rinfo", &rinfo); err != nil {
		return err
	}

	if err := AddOrUpdateFunc(&rinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", rinfo.Volname).Debug("failed to store rebalance details")
		return err
	}
	return nil
}

// AddOrUpdateFunc used to add or update the rebalance details
func AddOrUpdateFunc(r *RebalanceInfo) error {
	json, e := json.Marshal(r)
	if e != nil {
		log.WithField("error", e).Error("Failed to marshal the volinfo object")
		return e
	}

	_, e = store.Store.Put(context.TODO(), rebalancePrefix+r.Volname, string(json))
	if e != nil {
		log.WithError(e).Error("Couldn't add volume to store")
		return e
	}
	return nil
}
