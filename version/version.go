package version

import (
	"fmt"
	"runtime"

	flag "github.com/spf13/pflag"
)

// MaxOpVersion and APIVersion supported
const (
	MaxOpVersion = 40000
	APIVersion   = 1
)

// GlusterdVersion and GitSHA
var (
	GlusterdVersion = "4.0dev"
	GitSHA          = ""
)

func init() {
	flag.Bool("version", false, "Show the version information")
}

// DumpVersionInfo prints all version information
func DumpVersionInfo() {
	fmt.Printf("glusterd version: %s\n", GlusterdVersion)
	fmt.Printf("git SHA: %s\n", GitSHA)
	fmt.Printf("go version: %s\n", runtime.Version())
	fmt.Printf("go OS/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
