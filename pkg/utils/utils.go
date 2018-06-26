package utils

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/gluster/glusterd2/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// IsLocalAddress checks whether a given host/IP is local
// Does lookup only after string matching IP addresses
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

// ParseHostAndBrickPath parses the host & brick path out of req.Bricks list
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

// InitDir creates directory path and checks if files can be created in it.
// Returns error if path is not a directory or if directory doesn't have
// write permission.
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

// GetLocalIP will give local IP address of this node
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback then return it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
			if ipnet.IP.To16() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.ErrIPAddressNotFound
}

// GetFuncName returns the name of the passed function pointer
func GetFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// StringInSlice will return true if the given string is present in the
// list of strings provided. Will return false otherwise.
func StringInSlice(query string, list []string) bool {
	for _, s := range list {
		if s == query {
			return true
		}
	}
	return false
}

// IsAddressSame checks is two host addresses are same
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

// ExecuteCommandError represents command execution error
type ExecuteCommandError struct {
	ExitStatus int
	Errstr     string
	Err        error
}

func (e *ExecuteCommandError) Error() string {
	errstr := e.Errstr
	if errstr != "" {
		errstr = "; " + errstr
	}
	return e.Err.Error() + errstr
}

func execStderrCombined(err error, stderr *bytes.Buffer) error {
	if err == nil {
		return nil
	}

	execErr := ExecuteCommandError{
		ExitStatus: -1,
		Errstr:     stderr.String(),
		Err:        err,
	}

	exiterr, ok := err.(*exec.ExitError)
	if ok {
		status, ok := exiterr.Sys().(syscall.WaitStatus)
		if ok {
			execErr.ExitStatus = status.ExitStatus()
		}
	}

	return &execErr
}

// ExecuteCommandOutput runs the command and adds additional error information
func ExecuteCommandOutput(cmdName string, arg ...string) ([]byte, error) {
	cmd := exec.Command(cmdName, arg...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()

	if err != nil {
		return out, execStderrCombined(err, &stderr)
	}

	return out, nil
}

// ExecuteCommandRun runs the command and adds additional
// error information
func ExecuteCommandRun(cmdName string, arg ...string) error {
	cmd := exec.Command(cmdName, arg...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	return execStderrCombined(cmd.Run(), &stderr)
}
