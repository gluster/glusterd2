package volgen

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"

	config "github.com/spf13/viper"
)

const (
	thinArbiterOptionName = "replicate.thin-arbiter"
)

var varStrRE = regexp.MustCompile(`\{\{\s*(\S+)\s*\}\}`)

// UnknownVarStrErr is returned when a varstring is not found in the given map
type UnknownVarStrErr string

func boolify(value string) bool {
	switch value {
	case "on", "true", "enabled", "1":
		return true
	default:
		return false
	}
}

func (e UnknownVarStrErr) Error() string {
	return fmt.Sprintf("unknown variable string: %s", string(e))
}

func isVarStr(s string) bool {
	return varStrRE.MatchString(s)
}

type replacekey struct {
	key     string
	fullKey string
}

func varStr(s string) []replacekey {
	matches := varStrRE.FindAllString(s, -1)
	var strs []replacekey
	for _, m := range matches {
		mtrim := strings.Trim(m, "{} ")
		if mtrim == "" {
			continue
		}
		strs = append(strs, replacekey{key: mtrim, fullKey: m})
	}
	return strs
}

func varStrReplace(s string, vals map[string]string) (string, error) {
	if !isVarStr(s) {
		return s, nil
	}
	keys := varStr(s)
	for _, k := range keys {
		v, ok := vals[k.key]
		if !ok {
			return "", UnknownVarStrErr(k.key)
		}
		s = strings.Replace(s, k.fullKey, v, -1)
	}
	return s, nil
}

// SaveToFile saves the generated volfile content to given file path
func SaveToFile(filename string, content string) error {
	err := os.MkdirAll(path.Dir(filename), os.ModeDir|os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return ioutil.WriteFile(filename, []byte(content), 0644)
}

// DeleteFile deletes the given Volfile
func DeleteFile(volfileID string) error {
	err := os.Remove(volfileID + ".vol")
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ClusterVolfileToFile accepts temporary volinfo and mixes with the list of volinfo
// and generates the cluster level volfile
func ClusterVolfileToFile(v *volume.Volinfo, volfileID string, tmplName string) error {
	// Use temporary Volume info instead of Volume info from store
	clusterinfo, err := volume.GetVolumes(context.TODO())
	for idx, vinfo := range clusterinfo {
		if v != nil && vinfo.Name == v.Name {
			clusterinfo[idx] = v
			break
		}
	}

	tmpl, err := GetTemplateFromVolinfo(v, tmplName)
	if err != nil {
		return err
	}

	volfile, err := ClusterLevelVolfile(tmpl, clusterinfo)
	if err != nil {
		return err
	}

	filename := path.Join(config.GetString("localstatedir"), "volfiles", volfileID+".vol")
	return SaveToFile(filename, volfile)
}

// VolumeVolfileToFile generates Volume level volfile for the given template name
func VolumeVolfileToFile(volinfo *volume.Volinfo, volfileID string, tmplName string) error {
	tmpl, err := GetTemplateFromVolinfo(volinfo, tmplName)
	if err != nil {
		return err
	}

	volfile, err := VolumeLevelVolfile(tmpl, volinfo)
	if err != nil {
		return err
	}

	filename := path.Join(config.GetString("localstatedir"), "volfiles", volfileID+".vol")
	return SaveToFile(filename, volfile)
}

// BrickVolfileToFile generates Volume level volfile for the given template name
func BrickVolfileToFile(volinfo *volume.Volinfo, volfileID string, tmplName string, peerid string, brickPath string) error {
	tmpl, err := GetTemplateFromVolinfo(volinfo, tmplName)
	if err != nil {
		return err
	}

	volfile, err := BrickLevelVolfile(tmpl, volinfo, peerid, brickPath)
	if err != nil {
		return err
	}

	filename := path.Join(config.GetString("localstatedir"), "volfiles", volfileID+".vol")
	return SaveToFile(filename, volfile)
}

type stringMapBrick struct {
	StringMap map[string]string
}

type stringMapSubvol struct {
	StringMap map[string]string
	Bricks    []stringMapBrick
}

type stringMapVolume struct {
	StringMap map[string]string
	Subvols   []stringMapSubvol
}

// getExtraStringMaps prepares extra information which are required to replace
// var strings in xlator options
// Volume Level: {{ volume.decommissioned-bricks }}
// Subvol Level: {{ subvol.afr-pending-xattr }}
// Brick Level: {{ brick.index }}
func getExtraStringMaps(volinfo *volume.Volinfo) stringMapVolume {
	data := stringMapVolume{}
	data.Subvols = make([]stringMapSubvol, len(volinfo.Subvols))
	thinArbiterEnabled := false
	thinarbiter, exists := volinfo.Options[thinArbiterOptionName]

	if exists && thinarbiter != "" {
		thinArbiterEnabled = true
	}

	var decommissionedBricks []string
	clientIdx := 0

	for sidx, sv := range volinfo.Subvols {
		var afrPendingXattrs []string
		data.Subvols[sidx].Bricks = make([]stringMapBrick, len(sv.Bricks))

		for bidx, b := range sv.Bricks {
			data.Subvols[sidx].Bricks[bidx].StringMap = map[string]string{
				"brick.index": strconv.Itoa(clientIdx),
			}
			if b.Decommissioned {
				decommissionedBricks = append(
					decommissionedBricks,
					fmt.Sprintf("%s-client-%d", sv.Name, bidx),
				)
			}

			afrPendingXattrs = append(
				afrPendingXattrs,
				fmt.Sprintf("%s-client-%d", volinfo.Name, clientIdx),
			)
			clientIdx++
		}

		if thinArbiterEnabled {
			afrPendingXattrs = append(
				afrPendingXattrs,
				fmt.Sprintf("%s-ta-%d", volinfo.Name, clientIdx),
			)
			clientIdx++
		}
		data.Subvols[sidx].StringMap = map[string]string{
			"subvol.afr-pending-xattr": strings.Join(afrPendingXattrs, ","),
		}
	}

	data.StringMap = map[string]string{
		"volume.decommissioned-bricks": strings.Join(decommissionedBricks, " "),
	}

	return data
}
