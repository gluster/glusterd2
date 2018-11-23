package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/volume"
)

// Fops
var fops = []string{
	"NULL",
	"STAT",
	"READLINK",
	"MKNOD",
	"MKDIR",
	"UNLINK",
	"RMDIR",
	"SYMLINK",
	"RENAME",
	"LINK",
	"TRUNCATE",
	"OPEN",
	"READ",
	"WRITE",
	"STATFS",
	"FLUSH",
	"FSYNC", /* 16 */
	"SETXATTR",
	"GETXATTR",
	"REMOVEXATTR",
	"OPENDIR",
	"FSYNCDIR",
	"ACCESS",
	"CREATE",
	"FTRUNCATE",
	"FSTAT", /* 25 */
	"LK",
	"LOOKUP",
	"READDIR",
	"INODELK",
	"FINODELK",
	"ENTRYLK",
	"FENTRYLK",
	"XATTROP",
	"FXATTROP",
	"FGETXATTR",
	"FSETXATTR",
	"RCHECKSUM",
	"SETATTR",
	"FSETATTR",
	"READDIRP",
	"FORGET",
	"RELEASE",
	"RELEASEDIR",
	"GETSPEC",
	"FREMOVEXATTR",
	"FALLOCATE",
	"DISCARD",
	"ZEROFILL",
	"IPC",
	"SEEK",
	"LEASE",
	"COMPOUND",
	"GETACTIVELK",
	"SETACTIVELK",
	"PUT",
	"ICREATE",
	"NAMELINK",
	"MAXVALUE",
}

var profileSessionKeys = [...]string{"io-stats.count-fop-hits", "io-stats.latency-measurement"}

// BrickProfileInfo holds profile info of each brick
type BrickProfileInfo struct {
	BrickName       string   `json:"brick-name"`
	CumulativeStats StatType `json:"cumulative-stats,omitempty"`
	IntervalStats   StatType `json:"interval-stats,omitempty"`
}

// StatType contains profile info of cumulative/interval stats of a brick
type StatType struct {
	Duration             string                       `json:"duration"`
	DataRead             string                       `json:"data-read"`
	DataWrite            string                       `json:"data-write"`
	Interval             string                       `json:"interval"`
	PercentageAvgLatency float64                      `json:"percentage-avg-latency"`
	StatsInfo            map[string]map[string]string `json:"stat-info,omitempty"`
}

// getActiveProfileSession return true if there is any active volume profile session otherwise returns false.
func getActiveProfileSession(v *volume.Volinfo) bool {
	for _, key := range profileSessionKeys {
		value, ok := v.Options[key]
		if ok && value == "on" {
			continue
		}
		return false
	}
	return true
}
