package size

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Size represents unit to measure information size
type Size uint64

// Byte represents one byte of information
const Byte Size = 1

const (
	// KiB is multiple of unite Byte the binary prefix Ki represents 2^10
	KiB = 1024 * Byte
	// MiB is multiple of unite Byte the binary prefix Mi represents 2^20
	MiB = 1024 * KiB
	// GiB is multiple of unite Byte the binary prefix Gi represents 2^30
	GiB = 1024 * MiB
	// TiB is multiple of unite Byte the binary prefix Ti represents 2^40
	TiB = 1024 * GiB
	// PiB is multiple of unite Byte the binary prefix Pi represents 2^50
	PiB = 1024 * TiB
)

const (
	// KB is a multiple of the unit byte the prefix K represents 10^3
	KB = 1e3 * Byte
	// MB is a multiple of the unit byte the prefix M represents 10^6
	MB = 1e3 * KB
	// GB is a multiple of the unit byte the prefix G represents 10^9
	GB = 1e3 * MB
	// TB is a multiple of the unit byte the prefix T represents 10^12
	TB = 1e3 * GB
	// PB is a multiple of the unit byte the prefix T represents 10^15
	PB = 1e3 * TB
)

var sizeMultiple = map[string]Size{
	"B": Byte,

	"KB": KB,
	"MB": MB,
	"GB": GB,
	"TB": TB,
	"PB": PB,

	"K": KiB,
	"M": MiB,
	"G": GiB,
	"T": TiB,
	"P": PiB,

	"KiB": KiB,
	"MiB": MiB,
	"GiB": GiB,
	"TiB": TiB,
	"PiB": PiB,

	"Ki": KiB,
	"Mi": MiB,
	"Gi": GiB,
	"Ti": TiB,
	"Pi": PiB,
}

// Bytes returns number of bytes
func (s Size) Bytes() int64 { return int64(s) }

// KiloBytes returns numbers of KiloBytes in floating point
func (s Size) KiloBytes() float64 {
	kb := s / KB
	bytes := s % KB

	return float64(kb) + float64(bytes)/1e3
}

// MegaBytes returns numbers of MegaBytes in floating point
func (s Size) MegaBytes() float64 {
	mb := s / MB
	bytes := s % MB

	return float64(mb) + float64(bytes)/(1e6)
}

// GigaBytes returns number of GigaBytes in floating point
func (s Size) GigaBytes() float64 {
	gb := s / GB
	bytes := s % GB

	return float64(gb) + float64(bytes)/(1e9)
}

// TeraBytes returns number of TeraBytes in floating point
func (s Size) TeraBytes() float64 {
	tb := s / TB
	bytes := s % TB

	return float64(tb) + float64(bytes)/(1e12)
}

// KibiBytes returns number of KiB in floating point
func (s Size) KibiBytes() float64 {
	kib := s / KiB
	bytes := s % KiB

	return float64(kib) + float64(bytes)/1024
}

// MebiBytes returns number of MiB in floating point
func (s Size) MebiBytes() float64 {
	mib := s / MiB
	bytes := s % MiB

	return float64(mib) + float64(bytes)/(1024*1024)
}

// GibiBytes returns number of GiB in floating point
func (s Size) GibiBytes() float64 {
	gib := s / GiB
	bytes := s % GiB

	return float64(gib) + float64(bytes)/(1024*1024*1024)
}

// TebiBytes returns number of TiB in floating point
func (s Size) TebiBytes() float64 {
	tib := s / TiB
	bytes := s % TiB

	return float64(tib) + float64(bytes)/(1024*1024*1024*1024)
}

// String string representation of Size in form XXKiB/MiB/TiB/GiB/B
// TODO: support for string representation in other formats
func (s Size) String() string {

	if s >= TiB {
		return fmt.Sprintf("%.2fTiB", s.TebiBytes())
	}

	if s >= GiB {
		return fmt.Sprintf("%.2fGiB", s.GibiBytes())
	}

	if s >= MiB {
		return fmt.Sprintf("%.2fMiB", s.MebiBytes())
	}

	if s >= KiB {
		return fmt.Sprintf("%.2fKiB", s.KibiBytes())
	}

	return fmt.Sprintf("%d B", s)
}

var validSizePattern = regexp.MustCompile(
	`(\d+(\.\d+)?)([KMGTP]?i?B?)`,
)

// Parse parses a string representation of size and returns the Size value it represents.
// Supported formats are {PiB,TiB,GiB,MiB,KiB,PB,TB,GB,MB,KB,B,Pi,Ti,Gi,Mi,Ki}
func Parse(sizeStr string) (Size, error) {
	sizeStr = strings.Replace(sizeStr, " ", "", -1)

	if !validSizePattern.MatchString(sizeStr) {
		return 0, fmt.Errorf("size parse error: %s", sizeStr)
	}

	matches := validSizePattern.FindStringSubmatch(sizeStr)

	if len(matches) != 4 {
		return 0, fmt.Errorf("size parse error: %s invalid fields (%v)", sizeStr, matches)
	}

	if matches[0] != sizeStr {
		return 0, fmt.Errorf("size parse error: %s invalid fields (%v)", sizeStr, matches)
	}

	unit := matches[3]
	size := matches[1]

	multiplier, ok := sizeMultiple[unit]
	if !ok {
		return 0, fmt.Errorf("multiplier not found for unit: %s", unit)
	}

	val, err := strconv.ParseFloat(size, 64)
	if err != nil {
		return 0, err
	}

	return Size(float64(multiplier) * val), nil
}
