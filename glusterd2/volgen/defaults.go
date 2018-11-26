package volgen

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	config "github.com/spf13/viper"
)

// LoadDefaultTemplates loads the templates from the templates directory
// If templates not available, generates the default templates
func LoadDefaultTemplates() error {
	defaultTemplatesPath := path.Join(config.GetString("localstatedir"), "templates", "default.json")
	// If directory not exists, create the directory and then generate default templates
	_, err := os.Stat(defaultTemplatesPath)
	if os.IsNotExist(err) {
		content, err := json.MarshalIndent(namespaces[DefaultTemplateNamespace], "", "    ")
		if err != nil {
			return err
		}

		err = os.MkdirAll(path.Dir(defaultTemplatesPath), os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(defaultTemplatesPath, content, 0640)
	} else if err == nil {
		content, err := ioutil.ReadFile(defaultTemplatesPath)
		if err != nil {
			return err
		}
		var tmpls Templates
		err = json.Unmarshal(content, &tmpls)
		if err != nil {
			return err
		}
		namespaces[DefaultTemplateNamespace] = tmpls
		return nil
	}
	return err
}

func init() {
	tmpls := make(map[string]Template)
	// default brick template
	tmpls["brick"] = Template{
		Name:  "brick",
		Level: VolfileLevelBrick,
		Xlators: []Xlator{
			{
				Type: "protocol/server",
			},
			{
				Type:     "debug/io-stats",
				NameTmpl: "{{ brick.path }}",
			},
			{
				Type: "features/index",
			},
			{
				Type: "features/barrier",
			},
			{
				Type: "performance/io-threads",
			},
			{
				Type: "features/upcall",
			},
			{
				Type: "features/locks",
			},
			{
				Type: "features/access-control",
			},
			{
				Type: "features/bitrot-stub",
			},
			{
				Type:           "features/changelog",
				Disabled:       true,
				EnableByOption: true,
				Options: map[string]string{
					"capture-del-path": "on",
				},
			},
			{
				Type:     "features/arbiter",
				Disabled: true,
			},
			{
				Type: "storage/posix",
			},
		},
	}

	// default client template
	tmpls["client"] = Template{
		Name:  "fuse",
		Level: VolfileLevelVolume,
		Xlators: []Xlator{
			{
				Type:     "debug/io-stats",
				NameTmpl: "{{ volume.name }}",
			},
			{
				Type:           "features/read-only",
				Disabled:       true,
				EnableByOption: true,
			},
			{
				Type: "cluster/distribute",
			},
		},
		SubvolGraphXlators: []Xlator{
			{
				NameTmpl: "{{ subvol.name }}",
				TypeTmpl: "cluster/{{ subvol.type }}",
				Options: map[string]string{
					"afr-pending-xattr": "{{ subvol.afr-pending-xattr }}",
				},
			},
		},
		BrickGraphXlators: []Xlator{
			{
				Type:     "protocol/client",
				NameTmpl: "{{ subvol.name }}-client-{{ brick.index }}",
			},
		},
	}

	// default rebalance template
	tmpls["rebalance"] = Template{
		Name:  "rebalance",
		Level: VolfileLevelVolume,
		Xlators: []Xlator{
			{
				Type:     "debug/io-stats",
				NameTmpl: "{{ volume.name }}",
				Options: map[string]string{
					"log-level": "DEBUG",
				},
			},
			{
				Type: "cluster/distribute",
				Options: map[string]string{
					"decommissioned-bricks": "{{ volume.decommissioned-bricks }}",
				},
			},
		},
		SubvolGraphXlators: []Xlator{
			{
				NameTmpl: "{{ subvol.name }}",
				TypeTmpl: "cluster/{{ subvol.type }}",
				Options: map[string]string{
					"afr-pending-xattr": "{{ subvol.afr-pending-xattr }}",
				},
			},
		},
		BrickGraphXlators: []Xlator{
			{
				Type:     "protocol/client",
				NameTmpl: "{{ subvol.name }}-client-{{ brick.index }}",
			},
		},
	}

	// default glustershd template
	tmpls["glustershd"] = Template{
		Name:  "glustershd",
		Level: VolfileLevelCluster,
		Xlators: []Xlator{
			{
				Type:     "debug/io-stats",
				NameTmpl: "glustershd",
			},
		},
		SubvolGraphXlators: []Xlator{
			{
				TypeTmpl: "cluster/{{ subvol.type }}",
				Options: map[string]string{
					"iam-self-heal-daemon": "yes",
					"afr-pending-xattr":    "{{ subvol.afr-pending-xattr }}",
				},
			},
		},
		BrickGraphXlators: []Xlator{
			{
				Type: "protocol/client",
			},
		},
	}

	tmpls["bitd"] = Template{
		Name:  "bitd",
		Level: VolfileLevelCluster,
		Xlators: []Xlator{
			{
				Type:     "debug/io-stats",
				NameTmpl: "bitd",
			},
		},
		VolumeGraphXlators: []Xlator{
			{
				Type:            "features/bit-rot",
				NameTmpl:        "{{ volume.name }}",
				IgnoreOptions:   []string{"scrubber"},
				Disabled:        true,
				OnlyLocalBricks: true,
			},
		},
		SubvolGraphXlators: []Xlator{
			{
				TypeTmpl:        "cluster/{{ subvol.type }}",
				OnlyLocalBricks: true,
			},
		},
		BrickGraphXlators: []Xlator{
			{
				Type:            "protocol/client",
				OnlyLocalBricks: true,
			},
		},
	}

	tmpls["scrubd"] = Template{
		Name:  "scrubd",
		Level: VolfileLevelCluster,
		Xlators: []Xlator{
			{
				Type:     "debug/io-stats",
				NameTmpl: "scrub",
			},
		},
		VolumeGraphXlators: []Xlator{
			{
				Type:            "features/bit-rot",
				NameTmpl:        "{{ volume.name }}",
				OnlyLocalBricks: true,
				Disabled:        true,
			},
		},
		SubvolGraphXlators: []Xlator{
			{
				TypeTmpl:        "cluster/{{ subvol.type }}",
				OnlyLocalBricks: true,
			},
		},
		BrickGraphXlators: []Xlator{
			{
				Type:            "protocol/client",
				OnlyLocalBricks: true,
			},
		},
	}

	namespaces[DefaultTemplateNamespace] = tmpls
}
