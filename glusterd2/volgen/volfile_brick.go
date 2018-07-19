package volgen

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/pborman/uuid"
)

func generateBrickVolfile(volfile *Volfile, b *brick.Brickinfo, vol *volume.Volinfo, peerid uuid.UUID) {
	volfile.FileName = fmt.Sprintf("%s.%s.%s",
		vol.Name,
		b.PeerID,
		strings.Trim(strings.Replace(b.Path, "/", "-", -1), "-"),
	)
	last := volfile.RootEntry.
		Add("protocol/server", vol, b).
		Add("performance/decompounder", vol, b).SetName(b.Path).
		Add("debug/io-stats", vol, b).
		Add("features/sdfs", vol, b).
		Add("features/quota", vol, b).
		Add("features/index", vol, b).
		Add("features/barrier", vol, b).
		Add("features/marker", vol, b).
		Add("features/selinux", vol, b).
		Add("performance/io-threads", vol, b).
		Add("features/upcall", vol, b).
		Add("features/leases", vol, b).
		Add("features/read-only", vol, b).
		Add("features/worm", vol, b).
		Add("features/locks", vol, b).
		Add("features/access-control", vol, b).
		Add("features/bitrot-stub", vol, b).
		Add("features/changelog", vol, b).
		Add("features/changetimerecorder", vol, b).
		Add("features/trash", vol, b)

	if b.Type == brick.Arbiter {
		last = last.Add("features/arbiter", vol, b)
	}
	last.Add("storage/posix", vol, b)
}

func init() {
	registerBrickVolfile("brick", generateBrickVolfile, false)
}
