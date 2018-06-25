package cmd

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	validSizeFormat = regexp.MustCompile(`^([0-9]+)([GMKT])?$`)
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

func sizeToMb(value string) (uint64, error) {
	sizeParts := validSizeFormat.FindStringSubmatch(value)
	if len(sizeParts) == 0 {
		return 0, errors.New("invalid size format")
	}

	// If Size unit is specified as M/K/G/T, Default Size unit is M
	sizeUnit := "M"
	if len(sizeParts) == 3 {
		sizeUnit = sizeParts[2]
	}

	sizeValue, err := strconv.ParseUint(sizeParts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	var size uint64
	switch sizeUnit {
	case "K":
		size = sizeValue / 1024
	case "G":
		size = sizeValue * 1024
	case "T":
		size = sizeValue * 1024 * 1024
	default:
		size = sizeValue
	}
	return size, nil
}
