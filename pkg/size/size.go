package size

import (
	"errors"
	"fmt"
	"regexp"
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

// String string representation of Size in form XXKB/MB/TB/GB/Bytes
// TODO: support for string representation in XiB format
func (s Size) String() string {

	if s >= TB {
		return fmt.Sprintf("%.2fTB", s.TeraBytes())
	}

	if s >= GB {
		return fmt.Sprintf("%.2fGB", s.GigaBytes())
	}

	if s >= MB {
		return fmt.Sprintf("%.2fMB", s.MegaBytes())
	}

	if s >= KB {
		return fmt.Sprintf("%.2fKB", s.KiloBytes())
	}

	return fmt.Sprintf("%d Bytes", s)
}

// Parse parses a string representation of size and returns the Size value it represents.
// Supported formats are {TiB,GiB,MiB,KiB,TB,GB,MB,KB}
func Parse(s string) (Size, error) {
	var (
		count float64
		size  Size
		err   error
		regex = regexp.MustCompile(`^([\d.]+)([KMGT]i?B)$`)
	)

	s = strings.Replace(s, " ", "", -1)
	matches := regex.FindStringSubmatch(s)

	if len(matches) != 3 {
		return size, errors.New("invalid size format")
	}

	switch matches[2] {
	case "GiB":
		_, err = fmt.Sscanf(s, "%fGiB", &count)
		size = Size(count * float64(1*GiB))

	case "MiB":
		_, err = fmt.Sscanf(s, "%fMiB", &count)
		size = Size(count * float64(1*MiB))

	case "KiB":
		_, err = fmt.Sscanf(s, "%fKiB", &count)
		size = Size(count * float64(1*KiB))

	case "TiB":
		_, err = fmt.Sscanf(s, "%fTiB", &count)
		size = Size(count * float64(1*TiB))

	case "KB":
		_, err = fmt.Sscanf(s, "%fKB", &count)
		size = Size(count * float64(1*KB))

	case "MB":
		_, err = fmt.Sscanf(s, "%fMB", &count)
		size = Size(count * float64(1*MB))

	case "GB":
		_, err = fmt.Sscanf(s, "%fGB", &count)
		size = Size(count * float64(1*GB))

	case "TB":
		_, err = fmt.Sscanf(s, "%fTB", &count)
		size = Size(count * float64(1*TB))

	default:
		return 0, errors.New("can not parse to size")
	}

	return size, err
}
