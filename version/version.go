package version

import (
	"expvar"
	"fmt"
	"runtime"

	flag "github.com/spf13/pflag"
)

var (
	expVer = expvar.NewString("version")
)

// MaxOpVersion and APIVersion supported
const (
	MaxOpVersion = 50000
	APIVersion   = 1
)

// GlusterdVersion and GitSHA
// These are set as flags during build time. The current values are just placeholders
var (
	GlusterdVersion = ""
	GitSHA          = ""
)

func init() {
	flag.Bool("version", false, "Show the version information")
	expVer.Set(GlusterdVersion)
}

// DumpVersionInfo prints all version information
func DumpVersionInfo() {
	fmt.Printf("glusterd version: %s\n", GlusterdVersion)
	fmt.Printf("git SHA: %s\n", GitSHA)
	fmt.Printf("go version: %s\n", runtime.Version())
	fmt.Printf("go OS/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
