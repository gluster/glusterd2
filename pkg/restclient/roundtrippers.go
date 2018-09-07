package restclient

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// debugRoundTripper is a middleware which debugs each outgoing request.
// It will dump information about request passing through it.
type debugRoundTripper struct {
	rt http.RoundTripper
}

func newDebugRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &debugRoundTripper{rt}
}

// RoundTrip dumps the request and response of a single http transaction.
//
// It will dump each outgoing request in its HTTP/1.x wire representation and
// also It will dump the received response in its HTTP/1.x wire representation.
func (d *debugRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	reqCtx := newRequestCtx(req)
	log.WithFields(log.Fields{
		"method":  reqCtx.reqVerb,
		"url":     reqCtx.reqURL.String(),
		"headers": reqCtx.headers(),
	}).Debug("sending request")

	if reqDump, err := httputil.DumpRequestOut(req, true); err == nil {
		log.Debug("\n", dump(">", reqDump))
	}

	defer func(begin time.Time) {
		respDur := time.Since(begin)
		if err != nil {
			log.WithError(err).Debug("failed to connect to gd2 server")
			return
		}

		log.WithFields(log.Fields{
			"method":   reqCtx.reqVerb,
			"url":      reqCtx.reqURL.String(),
			"status":   resp.Status,
			"duration": respDur.String(),
		}).Debug("response received")

		if respDump, err := httputil.DumpResponse(resp, true); err == nil {
			log.Debug("\n", dump("<", respDump))
		}
	}(time.Now())

	return d.rt.RoundTrip(req)
}

type requestCtx struct {
	reqVerb   string
	reqURL    *url.URL
	reqHeader http.Header
}

func newRequestCtx(req *http.Request) *requestCtx {
	r := &requestCtx{
		reqHeader: make(http.Header),
		reqURL:    new(url.URL),
	}
	r.reqVerb = req.Method
	*r.reqURL = *req.URL
	r.reqHeader = req.Header
	return r
}

func (r *requestCtx) headers() string {
	var header string
	for key, vals := range r.reqHeader {
		for _, val := range vals {
			header += fmt.Sprintf(` %s:%s `, key, val)
		}
	}
	return header
}

func dump(prefix string, data []byte) string {
	var dmp string
	fields := strings.FieldsFunc(string(data), func(r rune) bool {
		return r == '\n'
	})
	for _, field := range fields {
		dmp += fmt.Sprint(prefix + " " + field + "\n")
	}
	return dmp
}
