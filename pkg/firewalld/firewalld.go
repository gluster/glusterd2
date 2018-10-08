package firewalld

import (
	"errors"
	"strconv"

	"github.com/godbus/dbus"
)

const (
	fInterface = "org.fedoraproject.FirewallD1"
	fObjPath   = "/org/fedoraproject/FirewallD1"
)

// Protocol type represents the type of protocol
type Protocol string

// Supported protocol types
const (
	ProtoTCP Protocol = "tcp"
	ProtoUDP Protocol = "udp"
)

var (
	dbusObj   dbus.BusObject
	isRunning bool
)

// AddPort removes a port from runtime configuration of firewalld.
// If zone is set to empty string, it is applied to the default zone.
func AddPort(zone string, port int, protocol Protocol) error {

	if port <= 0 || (protocol != ProtoTCP && protocol != ProtoUDP) {
		return errors.New("invalid port or protocol")
	}

	if dbusObj == nil || !isRunning {
		return nil
	}

	portStr := strconv.Itoa(port)

	return dbusObj.Call(fInterface+".zone.addPort", 0, zone, portStr, string(protocol), 0).Store(&zone)
}

// RemovePort removes a port from runtime configuration of firewalld.
// If zone is set to empty string, it is applied to the default zone.
func RemovePort(zone string, port int, protocol Protocol) error {

	if port <= 0 || (protocol != ProtoTCP && protocol != ProtoUDP) {
		return errors.New("invalid port or protocol")
	}

	if dbusObj == nil || !isRunning {
		return nil
	}

	portStr := strconv.Itoa(port)

	return dbusObj.Call(fInterface+".zone.removePort", 0, zone, portStr, string(protocol)).Store(&zone)
}

// Init initializes dbus connection and checks if firewalld is running.
func Init() error {

	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}

	// this can never fail
	dbusObj = conn.Object(fInterface, dbus.ObjectPath(fObjPath))

	var zone string
	if err := dbusObj.Call(fInterface+".getDefaultZone", 0).Store(&zone); err != nil {
		conn.Close()
		return err
	}

	_ = zone
	isRunning = true

	return nil
}
