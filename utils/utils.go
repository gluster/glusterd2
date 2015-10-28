package utils

// #include "limits.h"
import "C"

import (
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/gluster/glusterd2/errors"
)

// IsLocalAddress checks whether a given host/IP is local
func IsLocalAddress(host string) bool {
	found := false

	addrs, e := net.LookupHost(host)
	if e != nil {
		log.Error(e.Error())
		return found
	}
	for _, addr := range addrs {
		names, err := net.LookupAddr(addr)
		if err != nil {
			log.Error(e.Error())
			return found
		}
		for _, name := range names {
			if name == host ||
				finalHost == "localhost.localdomain" ||
				finalHost == "localhost" {
				found = true
				return found
			}
		}
	}
	return found
}

// ParseHostAndBrickPath parses the host & brick path out of req.Bricks list
func ParseHostAndBrickPath(brickPath string) (string, string) {
	i := strings.LastIndex(brickPath, ":")
	if i == -1 {
		log.Error("Invalid brick path, it should be in the form of host:path")
		return "", ""
	}
	hostname := brickPath[0:i]
	path := brickPath[i+1 : len(brickPath)]

	return hostname, path
}

//ValidateBrickPathLength validates the length of the brick path
func ValidateBrickPathLength(brickPath string) int {
	//TODO : Check whether PATH_MAX is compatible across all distros
	if len(filepath.Clean(brickPath)) >= syscall.PathMax {
		log.Error("brickpath is too long")
		return -1
	}
	return 0
}

//ValidateBrickSubDirLength validates the length of each sub directories under
//the brick path
func ValidateBrickSubDirLength(brickPath string) int {
	subdirs := strings.Split(brickPath, string(os.PathSeparator))
	// Iterate over the sub directories and validate that they don't breach
	//  _POSIX_PATH_MAX validation
	for _, subdir := range subdirs {
		if len(subdir) >= C._POSIX_PATH_MAX {
			log.Error("sub directory path %v is too long", subdir)
			return -1
		}
	}
	return 0
}

//GetDeviceID fetches the device id of the device containing the file/directory
func GetDeviceID(f os.FileInfo) (int, error) {
	s := f.Sys()
	switch s := s.(type) {
	case *syscall.Stat_t:
		return int(s.Dev), nil
	}
	return -1, errors.ErrDeviceIDNotFound
}

//ValidateBrickPathStats checks whether the brick directory can be created with
//certain validations like directory checks, whether directory is part of mount
//point etc
func ValidateBrickPathStats(brickPath string, host string, force bool) error {
	var created bool
	var rootStat, brickStat, parentStat os.FileInfo
	err := os.Mkdir(brickPath, 0666)
	if err != nil {
		if !os.IsExist(err) {
			log.WithFields(log.Fields{
				"host":  host,
				"brick": brickPath,
			}).Error("Failed to create brick - ", err.Error())
			return err
		}
	} else {
		created = true
	}
	brickStat, err = os.Lstat(brickPath)
	if err != nil {
		log.WithFields(log.Fields{
			"host":  host,
			"brick": brickPath,
		}).Error("Failed to stat on brick path - ", err.Error())
		return err
	}
	if !created && !brickStat.IsDir() {
		log.WithFields(log.Fields{
			"host":  host,
			"brick": brickPath,
		}).Error("brick path which is already present is not a directory")
		return err
	}

	rootStat, err = os.Lstat("/")
	if err != nil {
		log.Error("Failed to stat on / -", err.Error())
		return err
	}

	parentBrick := path.Dir(brickPath)
	parentStat, err = os.Lstat(parentBrick)
	if err != nil {
		log.WithFields(log.Fields{
			"host":        host,
			"brick":       brickPath,
			"parentBrick": parentBrick,
		}).Error("Failed to stat on parent of the brick path")
		return err
	}

	if !force {
		var parentDeviceID, rootDeviceID, brickDeviceID int
		var e error
		parentDeviceID, e = GetDeviceID(parentStat)
		if e != nil {
			log.WithFields(log.Fields{
				"host":  host,
				"brick": brickPath,
			}).Error("Failed to find the device id for parent of brick path")

			return err
		}
		rootDeviceID, e = GetDeviceID(rootStat)
		if e != nil {
			log.Error("Failed to find the device id of '/'")
			return err
		}
		brickDeviceID, e = GetDeviceID(brickStat)
		if e != nil {
			log.WithFields(log.Fields{
				"host":  host,
				"brick": brickPath,
			}).Error("Failed to find the device id of the brick")
			return err
		}
		if brickDeviceID != parentDeviceID {
			log.WithFields(log.Fields{
				"host":  host,
				"brick": brickPath,
			}).Error(errors.ErrBrickIsMountPoint.Error())
			return errors.ErrBrickIsMountPoint
		} else if parentDeviceID == rootDeviceID {
			log.WithFields(log.Fields{
				"host":  host,
				"brick": brickPath,
			}).Error(errors.ErrBrickUnderRootPartition.Error())
			return errors.ErrBrickUnderRootPartition
		}

	}

	return nil
}
