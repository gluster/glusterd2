package volgen

import (
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/pborman/uuid"
)

func nodesFromClusterInfo(clusterinfo []*volume.Volinfo) map[string]uuid.UUID {
	set := make(map[string]uuid.UUID)
	for _, vol := range clusterinfo {
		for _, subvol := range vol.Subvols {
			for _, brick := range subvol.Bricks {
				if _, ok := set[brick.PeerID.String()]; !ok {
					set[brick.PeerID.String()] = brick.PeerID
				}
			}
		}
	}
	return set
}

type extrainfo struct {
	StringMaps map[string]map[string]string
	Options    map[string]string
}

func generateClusterLevelVolfiles(clusterinfo []*volume.Volinfo, xopts *map[string]extrainfo) error {
	for _, cvf := range clusterVolfiles {
		if cvf.nodeLevel {
			for _, peerid := range nodesFromClusterInfo(clusterinfo) {
				volfile := New(cvf.name)
				cvf.fn.(clusterVolfileFunc)(volfile, clusterinfo, peerid)
				volfiledata, err := volfile.Generate("", xopts)
				if err != nil {
					return err
				}
				err = save(peerid.String()+"-"+volfile.FileName, volfiledata)
				if err != nil {
					return err
				}
			}
		} else {
			volfile := New(cvf.name)
			cvf.fn.(clusterVolfileFunc)(volfile, clusterinfo, nil)
			volfiledata, err := volfile.Generate("", xopts)
			if err != nil {
				return err
			}
			err = save(volfile.FileName, volfiledata)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateVolumeLevelVolfiles(clusterinfo []*volume.Volinfo, xopts *map[string]extrainfo) error {
	for _, volinfo := range clusterinfo {
		for _, vvf := range volumeVolfiles {
			if vvf.nodeLevel {
				for _, peerid := range volinfo.Nodes() {
					volfile := New(vvf.name)
					vvf.fn.(volumeVolfileFunc)(volfile, volinfo, peerid)
					volfiledata, err := volfile.Generate("", xopts)
					if err != nil {
						return err
					}
					err = save(peerid.String()+"-"+volfile.FileName, volfiledata)
					if err != nil {
						return err
					}
				}
			} else {
				volfile := New(vvf.name)
				vvf.fn.(volumeVolfileFunc)(volfile, volinfo, nil)
				volfiledata, err := volfile.Generate("", xopts)
				if err != nil {
					return err
				}
				err = save(volfile.FileName, volfiledata)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func generateBrickLevelVolfiles(clusterinfo []*volume.Volinfo, xopts *map[string]extrainfo) error {
	for _, volinfo := range clusterinfo {
		nodes := volinfo.Nodes()
		for _, brick := range volinfo.GetBricks() {
			for _, bvf := range brickVolfiles {
				if bvf.nodeLevel {
					for _, peerid := range nodes {
						volfile := New(bvf.name)
						bvf.fn.(brickVolfileFunc)(volfile, &brick, volinfo, peerid)
						volfiledata, err := volfile.Generate("", xopts)
						if err != nil {
							return err
						}
						err = save(peerid.String()+"-"+volfile.FileName, volfiledata)
						if err != nil {
							return err
						}
					}
				} else {
					volfile := New(bvf.name)
					bvf.fn.(brickVolfileFunc)(volfile, &brick, volinfo, nil)
					volfiledata, err := volfile.Generate("", xopts)
					if err != nil {
						return err
					}
					err = save(volfile.FileName, volfiledata)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// Generate generates all the volfiles(Cluster/Volume/Brick)
func Generate() error {
	clusterinfo, err := volume.GetVolumes()
	if err != nil {
		return err
	}

	snapClusterInfo, err := snapshot.GetSnapshotVolumes()
	if err != nil {
		return err
	}
	clusterinfo = append(clusterinfo, snapClusterInfo...)
	err = generateVolfiles(clusterinfo)
	if err != nil {
		return err
	}

	return nil
}
func generateVolfiles(clusterinfo []*volume.Volinfo) error {
	var xopts = make(map[string]extrainfo)
	for _, vol := range clusterinfo {
		data := make(map[string]map[string]string)
		data[vol.ID.String()] = vol.StringMap()
		for _, subvol := range vol.Subvols {
			for _, b := range subvol.Bricks {
				data[vol.ID.String()+"."+b.ID.String()] = utils.MergeStringMaps(vol.StringMap(), b.StringMap())
			}
		}
		xopts[vol.ID.String()] = extrainfo{data, vol.Options}
	}

	// TODO: Note Start time and add metrics

	// Generate/Regenerate Cluster Level Volfiles
	err := generateClusterLevelVolfiles(clusterinfo, &xopts)
	if err != nil {
		return err
	}

	// Generate/Regenerate Volume Level Volfiles
	err = generateVolumeLevelVolfiles(clusterinfo, &xopts)
	if err != nil {
		return err
	}

	// Generate/Regenerate Brick Level Volfiles
	err = generateBrickLevelVolfiles(clusterinfo, &xopts)
	if err != nil {
		return err
	}
	return nil

}
