// Created by cgo - DO NOT EDIT

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:1
package utils

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:6
import (
	"net"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	testXattr     = "trusted.glusterfs.test"
	volumeIDXattr = "trusted.glusterfs.volume-id"
	gfidXattr     = "trusted.gfid"
)

var (
//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:31
	PathMax = unix.PathMax

	Removexattr = unix.Removexattr

	Setxattr = unix.Setxattr

	Getxattr = unix.Getxattr
)

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:41
const PosixPathMax = _Ciconst__POSIX_PATH_MAX

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:45
func IsLocalAddress(address string) (bool, error) {
	var host string

	host, _, _ = net.SplitHostPort(address)
	if host == "" {
		host = address
	}

	localNames := []string{"127.0.0.1", "localhost", "::1"}
	for _, name := range localNames {
		if host == name {
			return true, nil
		}
	}

	laddrs, e := net.InterfaceAddrs()
	if e != nil {
		return false, e
	}
	var lips []net.IP
	for _, laddr := range laddrs {
		lipa := laddr.(*net.IPNet)
		lips = append(lips, lipa.IP)
	}

	for _, ip := range lips {
		if host == ip.String() {
			return true, nil
		}
	}

	rips, e := net.LookupIP(host)
	if e != nil {
		return false, e
	}
	for _, rip := range rips {
		for _, lip := range lips {
			if lip.Equal(rip) {
				return true, nil
			}
		}
	}
	return false, nil
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:91
func ParseHostAndBrickPath(brickPath string) (string, string, error) {
	i := strings.LastIndex(brickPath, ":")
	if i == -1 {
		log.WithField("brick", brickPath).Error(errors.ErrInvalidBrickPath.Error())
		return "", "", errors.ErrInvalidBrickPath
	}
	hostname := brickPath[0:i]
	path := brickPath[i+1:]

	return hostname, path, nil
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:104
func ValidateBrickPathLength(brickPath string) error {

	if len(filepath.Clean(brickPath)) >= PathMax {
		log.WithField("brick", brickPath).Error(errors.ErrBrickPathTooLong.Error())
		return errors.ErrBrickPathTooLong
	}
	return nil
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:115
func ValidateBrickSubDirLength(brickPath string) error {
	subdirs := strings.Split(brickPath, string(os.PathSeparator))

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:119
	for _, subdir := range subdirs {
		if len(subdir) >= PosixPathMax {
			log.WithField("subdir", subdir).Error("sub directory path is too long")
			return errors.ErrSubDirPathTooLong
		}
	}
	return nil
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:129
func GetDeviceID(f os.FileInfo) (int, error) {
	s := f.Sys()
	switch s := s.(type) {

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:134
	case *syscall.Stat_t:
		return int(s.Dev), nil
	}
	return -1, errors.ErrDeviceIDNotFound
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:143
func ValidateBrickPathStats(brickPath string, force bool) error {
	var created bool
	var rootStat, brickStat, parentStat os.FileInfo
	err := os.MkdirAll(brickPath, os.ModeDir|os.ModePerm)
	if err != nil {
		if !os.IsExist(err) {
			log.WithFields(log.Fields{
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
			"brick": brickPath,
		}).Error("Failed to stat on brick path - ", err.Error())
		return err
	}
	if !created && !brickStat.IsDir() {
		log.WithFields(log.Fields{
			"brick": brickPath,
		}).Error("brick path which is already present is not a directory")
		return errors.ErrBrickNotDirectory
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
				"brick": brickPath,
			}).Error("Failed to find the device id of the brick")
			return err
		}
		if brickDeviceID != parentDeviceID {
			log.WithFields(log.Fields{
				"brick": brickPath,
			}).Error(errors.ErrBrickIsMountPoint.Error())
			return errors.ErrBrickIsMountPoint
		} else if parentDeviceID == rootDeviceID {
			log.WithFields(log.Fields{
				"brick": brickPath,
			}).Error(errors.ErrBrickUnderRootPartition.Error())
			return errors.ErrBrickUnderRootPartition
		}

	}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:225
	if err := os.MkdirAll(filepath.Join(brickPath, ".glusterfs", "indices"), os.ModeDir|os.ModePerm); err != nil {
		log.WithError(err).Error("failed to create .glusterfs/indices directory")
		return err
	}

	return nil
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:236
func ValidateXattrSupport(brickPath string, volid uuid.UUID, force bool) error {
	var err error
	err = Setxattr(brickPath, "trusted.glusterfs.test", []byte("working"), 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error(),
			"brickPath": brickPath,
			"xattr":     testXattr}).Error("setxattr failed")
		return err
	}
	err = Removexattr(brickPath, "trusted.glusterfs.test")
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error(),
			"brickPath": brickPath,
			"xattr":     testXattr}).Error("removexattr failed")
		return err
	}
	if !force {
		if isBrickPathAlreadyInUse(brickPath) {
			log.WithFields(log.Fields{
				"brickPath": brickPath}).Error(errors.ErrBrickPathAlreadyInUse.Error())
			return errors.ErrBrickPathAlreadyInUse
		}
	}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:261
	err = Setxattr(brickPath, volumeIDXattr, []byte(volid), 0)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error(),
			"brickPath": brickPath,
			"xattr":     volumeIDXattr}).Error("setxattr failed")
		return err
	}

	return nil
}

func isBrickPathAlreadyInUse(brickPath string) bool {
	keys := []string{gfidXattr, volumeIDXattr}
	var p string
	var buf []byte
	p = brickPath
	for ; p != "/"; p = path.Dir(p) {
		for _, key := range keys {
			size, err := Getxattr(p, key, buf)
			if err != nil {
				return false
			} else if size > 0 {
				return true
			} else {
				return false
			}

		}
	}
	return false
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:296
func InitDir(path string) error {

	if err := os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
		log.WithError(err).WithField("path", path).Debug(
			"failed to create directory")
		return err
	}

	if err := unix.Access(path, unix.W_OK); err != nil {
		log.WithError(err).WithField("path", path).Debug(
			"directory does not have write permission")
		return err
	}

	return nil
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:314
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {

		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.ErrIPAddressNotFound
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:332
func GetFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:338
func StringInSlice(query string, list []string) bool {
	for _, s := range list {
		if s == query {
			return true
		}
	}
	return false
}

//line /home/kaushal/go/src/github.com/gluster/glusterd2/pkg/utils/utils.go:348
func IsAddressSame(host1, host2 string) bool {

	if host1 == host2 {
		return true
	}

	addrs1, err := net.LookupHost(host1)
	if err != nil {
		return false
	}

	addrs2, err := net.LookupHost(host2)
	if err != nil {
		return false
	}

	for _, a := range addrs1 {
		if StringInSlice(a, addrs2) {
			return true
		}
	}

	return false
}
