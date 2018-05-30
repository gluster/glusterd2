//Package snapshot that contains struct for snapshot
package snapshot

import "github.com/gluster/glusterd2/glusterd2/volume"

//Snapinfo is used to represent a snapshot
type Snapinfo struct {
	SnapVolinfo  volume.Volinfo
	ParentVolume string
	Description  string
	OptionChange map[string]string
}
