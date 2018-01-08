package eventlistener

import (
	log "github.com/sirupsen/logrus"
	"net"
)

// EventListener is server for listening to UDP messages
type EventListener struct {
	connection *net.UDPConn
	stopCh     chan struct{}
}

// New Initializes event listener
func New() *EventListener {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":24009")
	if err != nil {
		log.WithError(err).Error("UDP address resolution failed")
	}

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		log.WithError(err).Error("Error getting server connection")
	}

	eventlistener := &EventListener{
		connection: ServerConn,
		stopCh:     make(chan struct{}),
	}

	log.Info("Initialized Event listener")

	return eventlistener
}

// Serve will start accepting UDP messages on udp port
// 24009.
func (e *EventListener) Serve() {

	buf := make([]byte, 1024)
	log.Info("Started Event listener")

	for {
		select {
		case <-e.stopCh:
			e.connection.Close()
			log.Info("Stopped Event listener")
			return
		default:
		}

		size, addr, err := e.connection.ReadFromUDP(buf)
		if err != nil {
			log.WithError(err).Error("Error while reading UDP message")
		}

		handleMessage(string(buf[0:size]), addr, err)

	}
}

// Stop stops the UDP server
func (e *EventListener) Stop() {
	close(e.stopCh)
}
