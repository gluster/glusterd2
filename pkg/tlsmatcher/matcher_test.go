package tlsmatcher

import (
	"bytes"
	"testing"
)

func TestTLS10(t *testing.T) {
	var cases = []struct {
		header []byte
		result bool
	}{
		{tls10Header[:], true},
		{tls11Header[:], false},
		{[]byte{20, 3, 1}, false},
		{[]byte{0, 0, 0}, false},
	}

	for _, c := range cases {
		r := bytes.NewReader(c.header)
		if TLS10(r) != c.result {
			t.Fail()
		}
	}
}

func TestTLS11(t *testing.T) {
	var cases = []struct {
		header []byte
		result bool
	}{
		{tls11Header[:], true},
		{tls12Header[:], false},
		{[]byte{20, 3, 2}, false},
		{[]byte{0, 0, 0}, false},
	}

	for _, c := range cases {
		r := bytes.NewReader(c.header)
		if TLS11(r) != c.result {
			t.Fail()
		}
	}
}

func TestTLS12(t *testing.T) {
	var cases = []struct {
		header []byte
		result bool
	}{
		{tls12Header[:], true},
		{tls10Header[:], false},
		{[]byte{20, 3, 3}, false},
		{[]byte{0, 0, 0}, false},
	}

	for _, c := range cases {
		r := bytes.NewReader(c.header)
		if TLS12(r) != c.result {
			t.Fail()
		}
	}
}
