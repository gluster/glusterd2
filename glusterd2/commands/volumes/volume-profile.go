package volumecommands

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/servers/sunrpc/dict"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/gorilla/mux"
)

type keyType int8

const (
	cumulativeType keyType = iota
	intervalType
)

func registerVolProfileStepFuncs() {
	transaction.RegisterStepFunc(txnVolumeProfile, "volume.Profile")
}

func volumeProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]
	option := mux.Vars(r)["option"]

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	if volinfo.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "volume must be in start state")
		return
	}

	if !getActiveProfileSession(volinfo) {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "there are no active profile sessions running")
		return

	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "volume.Profile",
			Nodes:  volinfo.Nodes(),
		},
	}

	if err := txn.Ctx.Set("option", option); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volname", volname).Error("transaction to profile volume failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	var volumeProfileInfo []BrickProfileInfo
	for _, node := range volinfo.Nodes() {
		var nodeResult []map[string]string
		err := txn.Ctx.GetNodeResult(node, "node-result", &nodeResult)
		if err != nil {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
			return
		}
		// Get profile Info Array for  each nodes of a volume
		for brickResult := range nodeResult {
			var brickProfileInfo BrickProfileInfo
			brickProfileInfo.BrickName = fmt.Sprintf("%s:%s", node, nodeResult[brickResult]["brick"])
			brickProfileInfo.CumulativeStats.Interval = nodeResult[brickResult]["cumulative"]
			brickProfileInfo.IntervalStats.Interval = nodeResult[brickResult]["interval"]
			// Store stats for each fop in cumulative stats
			brickProfileInfo.CumulativeStats.StatsInfo = make(map[string]map[string]string)
			// Store stats for each fop in Interval Stats
			brickProfileInfo.IntervalStats.StatsInfo = make(map[string]map[string]string)
			// Iterate over each brick info of a node
			for key, value := range nodeResult[brickResult] {
				// Assures if key is a part of cumulative stats and whether cumulative stats are present in the  profile info or not
				if strings.HasPrefix(key, nodeResult[brickResult]["cumulative"]) && nodeResult[brickResult]["cumulative"] != "" {
					brickProfileInfo.CumulativeStats = populateStatsWithFop(brickProfileInfo.CumulativeStats, key, value, cumulativeType)
				} else if strings.HasPrefix(key, nodeResult[brickResult]["interval"]) && nodeResult[brickResult]["interval"] != "" {
					brickProfileInfo.IntervalStats = populateStatsWithFop(brickProfileInfo.IntervalStats, key, value, intervalType)
				}
			}

			brickProfileInfo.CumulativeStats = calculatePercentageLatencyForEachFop(brickProfileInfo.CumulativeStats)

			brickProfileInfo.IntervalStats = calculatePercentageLatencyForEachFop(brickProfileInfo.IntervalStats)

			// Append each brick's profile Info in an array and return the array
			volumeProfileInfo = append(volumeProfileInfo, brickProfileInfo)
		}
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, &volumeProfileInfo)
}

// Calculate Percentage latency for each fop in cumulative/interval stats
func calculatePercentageLatencyForEachFop(stats StatType) StatType {
	// Calculate sum of (hits * avgLatency) of all fops in Cumulative Stats. Used Later to calculate
	// %-Latency for each FOP.
	var tmpPercentageAvgLatency float64
	for key := range stats.StatsInfo {
		tmpAvgLatency, _ := strconv.ParseFloat(stats.StatsInfo[key]["avglatency"], 64)
		tmpHits, _ := strconv.ParseFloat(stats.StatsInfo[key]["hits"], 64)
		tmpPercentageAvgLatency += tmpAvgLatency * tmpHits
	}
	// Calculate %-Latency for each fop
	// %-Avg-Latency  for one fop = 100 * (Hits for that fop * AvgLatency for that fop ) / (sum of hits * avgLatency of all fop)
	for key := range stats.StatsInfo {
		tmpAvgLatency, _ := strconv.ParseFloat(stats.StatsInfo[key]["avglatency"], 64)
		tmpHits, _ := strconv.ParseFloat(stats.StatsInfo[key]["hits"], 64)
		tmpPercentageAvgLatencyForEachFop := 100 * ((tmpAvgLatency * tmpHits) / tmpPercentageAvgLatency)
		stats.StatsInfo[key]["%-latency"] = fmt.Sprintf("%f", tmpPercentageAvgLatencyForEachFop)
	}
	stats.PercentageAvgLatency = tmpPercentageAvgLatency

	return stats
}

// populateStatsWithFop populates both interval/cumulative stats with fops and its profile info
func populateStatsWithFop(stats StatType, key string, value string, ktype keyType) StatType {
	var fop string
	var k string
	// Decode the fop and stat for the given key
	switch ktype {
	case cumulativeType:
		fop, k = decodeCumulativeKey(key)
		break
	case intervalType:
		fop, k = decodeIntervalKey(key)
		break
	}
	if fop != "" && fop != "NULL" {
		if _, ok := stats.StatsInfo[fop]; ok {
			stats.StatsInfo[fop][k] = value
		} else {
			// Create new map with fop as key if this particular Fop is encountered for the first time
			stats.StatsInfo[fop] = make(map[string]string)
			stats.StatsInfo[fop][k] = value
		}
	} else {
		if k == "read" {
			stats.DataRead = value
		} else if k == "write" {
			stats.DataWrite = value
		} else if k == "duration" {
			stats.Duration = value
		}
	}
	return stats
}

// Decode key for fop and stat. keys starting with -1 belong to cumulative stat
// Eg: -1-12-maxlatency fop : 12        stat: maxlatency
// -1-duration          fop : ""        stat: duration
func decodeCumulativeKey(key string) (string, string) {
	var fop string
	var k string
	s := strings.Split(key, "-")
	if len(s) == 4 {
		k = s[len(s)-1]
		index, _ := strconv.Atoi(s[2])
		fop = fops[index]
	} else if len(s) == 3 {
		k = s[len(s)-1]
		fop = ""
	}
	return fop, k
}

// Decode key for fop and stat for keys belonging to interval stat
// Eg: 12-12-maxlatency  IntervalNo.: 12  fop : 12        stat: maxlatency
// 12-duration           IntervalNo.: 12  fop : ""        stat: duration
func decodeIntervalKey(key string) (string, string) {
	var fop string
	var k string
	s := strings.Split(key, "-")
	if len(s) == 3 {
		k = s[len(s)-1]
		index, _ := strconv.Atoi(s[1])
		fop = fops[index]
	} else if len(s) == 2 {
		k = s[len(s)-1]
		fop = ""
	}
	return fop, k
}

func txnVolumeProfile(c transaction.TxnCtx) error {
	var volinfo volume.Volinfo
	var nodeProfileInfo []map[string]string
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	var option string
	if err := c.Get("option", &option); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		brickDaemon, err := brick.NewGlusterfsd(b)
		if err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.String()).Error("failed to inittiate brick daemon")
			return err
		}

		c.Logger().WithField(
			"volume", volinfo.Name).Info("Starting volume profile operation")

		client, err := daemon.GetRPCClient(brickDaemon)
		if err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.String()).Error("failed to connect to brick, aborting volume profile operation")
			return err
		}
		reqDict := make(map[string]string)
		switch option {
		case "info":
			reqDict["peek"] = "0"
			reqDict["op"] = "3"
			reqDict["info-op"] = "1"
			reqDict["originator_uuid"] = gdctx.MyUUID.String()
			break
		case "info-peek":
			reqDict["peek"] = "1"
			reqDict["op"] = "3"
			reqDict["info-op"] = "1"
			break
		case "info-incremental":
			reqDict["peek"] = "0"
			reqDict["op"] = "3"
			reqDict["info-op"] = "2"
			break
		case "info-incremental-peek":
			reqDict["peek"] = "1"
			reqDict["op"] = "3"
			reqDict["info-op"] = "2"
			break
		case "info-cumulative":
			reqDict["peek"] = "0"
			reqDict["op"] = "3"
			reqDict["info-op"] = "3"
			break
		case "info-clear":
			reqDict["peek"] = "0"
			reqDict["op"] = "3"
			reqDict["info-op"] = "4"
			break
		default:
			return fmt.Errorf("%s is  not a valid operation", option)
		}

		reqDict["volname"] = volinfo.Name
		reqDict["vol-id"] = volinfo.ID.String()
		req := &brick.GfBrickOpReq{
			Name: b.Path,
			Op:   int(brick.OpBrickXlatorInfo),
		}
		req.Input, err = dict.Serialize(reqDict)
		if err != nil {
			c.Logger().WithError(err).WithField(
				"reqDict", reqDict).Error("failed to convert map to slice of bytes")
			return err
		}
		var rsp brick.GfBrickOpRsp
		err = client.Call("Brick.OpBrickXlatorInfo", req, &rsp)
		if err != nil || rsp.OpRet != 0 {
			c.Logger().WithError(err).WithField(
				"brick", b.String()).Error("failed to send volume profile RPC")
			return err
		}

		output, err := dict.Unserialize(rsp.Output)
		if err != nil {
			return errors.New("error unserializing the output")
		}
		output["brick"] = b.Path
		nodeProfileInfo = append(nodeProfileInfo, output)
	}
	c.SetNodeResult(gdctx.MyUUID, "node-result", &nodeProfileInfo)

	return nil
}
