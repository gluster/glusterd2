// Package tlsmatcher implements github.com/cockroachdb/cmux.Matcher for TLS
// connections
//
// Ref: https://en.wikipedia.org/wiki/Transport_Layer_Security#TLS_record
// The general format for TLS record is as follows,
//
//	  Record type (1byte)
//	  |    TLS version major/minor (2bytes)
//	  |    |         2b payload length (2bytes)
//	  |    |         |
//	+----+----+----+----+----+--------
//	|    |    |    |    |    |
//	|    |  TLS Header  |    |  TLS Payload
//	+----+----+----+----+----+--------
//
// For cmux matching we are only interested in the record type and TLS version.
// We only need to match incoming handshake requests, ie record type 22, which
// is the first incoming TLS request of a new connection. Cmux will launch the
// relevant connection handler, which handles all further requests for the
// connection.
package tlsmatcher

import (
	"io"
)

type tlsHeader [3]byte

var (
	tls10Header = tlsHeader{22, 3, 1}
	tls11Header = tlsHeader{22, 3, 2}
	tls12Header = tlsHeader{22, 3, 3}
)

func matchTLS(r io.Reader, h tlsHeader) bool {
	var buf tlsHeader

	_, _ = io.ReadFull(r, buf[:])

	return buf == h
}

// TLS10 matches incoming TLSv1.0 connections
func TLS10(r io.Reader) bool {
	return matchTLS(r, tls10Header)
}

// TLS11 matches incoming TLSv1.1 connections
func TLS11(r io.Reader) bool {
	return matchTLS(r, tls11Header)
}

// TLS12 matches incoming TLSv1.2 connections
func TLS12(r io.Reader) bool {
	return matchTLS(r, tls12Header)
}
