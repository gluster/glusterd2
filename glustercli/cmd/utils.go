package cmd

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/pkg/utils"
)

var (
	validSizeFormat = regexp.MustCompile(`^([0-9]+)\s*([kKmMgGtT]?[iI]?[bB]?)$`)
)

func formatBoolYesNo(value bool) string {
	if value == true {
		return "yes"
	}
	return "no"
}

func formatPID(pid int) string {
	if pid == 0 {
		return ""
	}
	return strconv.Itoa(pid)
}

func sizeToBytes(value string) (uint64, error) {
	if value == "" {
		return 0, nil
	}
	sizeParts := validSizeFormat.FindStringSubmatch(value)
	if len(sizeParts) == 0 {
		return 0, errors.New("invalid size format")
	}
	// If Size unit is specified as M/K/G/T, Default Size unit is bytes
	var sizeUnit string
	if len(sizeParts) == 3 {
		sizeUnit = sizeParts[2]
	}

	sizeValue, err := strconv.ParseUint(sizeParts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	var size uint64
	switch strings.ToLower(sizeUnit) {
	case "k", "kib":
		size = sizeValue * utils.KiB
	case "kb":
		size = sizeValue * utils.KB
	case "m", "mib":
		size = sizeValue * utils.MiB
	case "mb":
		size = sizeValue * utils.MB
	case "g", "gib":
		size = sizeValue * utils.GiB
	case "gb":
		size = sizeValue * utils.GB
	case "t", "tib":
		size = sizeValue * utils.TiB
	case "tb":
		size = sizeValue * utils.TB
	case "b", "":
		size = sizeValue
	default:
		return 0, fmt.Errorf("invalid size unit specified %s", sizeUnit)
	}
	return size, nil
}

// logn is used to find the unit size the given bytes belongs to.
// 1024 will return 1
// 1048576 will return 2 and so on
func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

// humanReadable converts size given in bytes into a more human readable unit
//
// humanReadable(1024) returns 1.0 KiB
// humanReadable(1536) returns 1.5 KiB
// humanReadable(1048576) returns 1.0 MiB
func humanReadable(value uint64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

	// If less than 1024 we return it as such
	if value < 1024 {
		return fmt.Sprintf("%.1f B", float64(value))
	}
	e := math.Floor(logn(float64(value), 1024))
	suffix := units[int(e)]
	size := math.Floor(float64(value)/math.Pow(1024, e)*10+0.5) / 10
	return fmt.Sprintf("%.1f %s", size, suffix)
}

func readString(prompt string, args ...interface{}) string {
	var s string
	fmt.Printf(prompt, args...)
	fmt.Scanln(&s)
	return s
}

// PromptConfirm prompts for confirmation
func PromptConfirm(prompt string, args ...interface{}) bool {
	switch strings.ToLower(readString(prompt, args...)) {
	case "yes", "y":
		return true
	default:
		return false
	}
}
