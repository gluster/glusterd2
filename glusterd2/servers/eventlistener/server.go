package eventlistener

import (
	"net"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

const packetBufferSize = 1024

// EventListener is server for listening to UDP messages
type EventListener struct {
	udpConn *net.UDPConn
	stopCh  chan struct{}
}

// New initializes event listener
func New() *EventListener {

	udpAddr, err := net.ResolveUDPAddr("udp", config.GetString("clientaddress"))
	if err != nil {
		// TODO: Bubble up error instead of Fatal()
		log.WithError(err).WithField("address",
			config.GetString("clientaddress")).Fatal("UDP address resolution failed")
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		// TODO: Bubble up error instead of Fatal()
		log.WithError(err).Fatal("ListenUDP() failed")
	}

	return &EventListener{
		udpConn: udpConn,
		stopCh:  make(chan struct{}),
	}
}

// Serve will start accepting UDP messages.
func (l *EventListener) Serve() {

	log.WithFields(log.Fields{
		"address":   config.GetString("clientaddress"),
		"transport": "udp"}).Info("started event listener")

	buf := make([]byte, packetBufferSize)
	for {
		select {
		case <-l.stopCh:
			log.Info("stopped event listener")
			return
		default:
		}

		size, addr, err := l.udpConn.ReadFromUDP(buf)
		if err != nil {
			log.WithError(err).Error("Error while reading UDP message")
			continue
		}

		handleMessage(string(buf[:size]), addr)
	}
}

// Stop stops the UDP server
func (l *EventListener) Stop() {
	close(l.stopCh)
	l.udpConn.Close()
}
