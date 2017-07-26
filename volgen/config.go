package volgen

import (
	"path"

	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

const (
	templateDirOpt = "templatesdir"
	templateDir    = "templates"
)

// InitFlags intializes the commandline options for volgen
func InitFlags() {
	flag.String(templateDirOpt, "", "Directory to search for templates. (default: workdir/templates)")
}

// SetDefaults sets the default values for the volgen commandline options
func SetDefaults() {
	td := config.GetString(templateDirOpt)
	if td == "" {
		wd := config.GetString("workdir")
		config.SetDefault(templateDirOpt, path.Join(wd, templateDir))
	}
}
