package utils

//Based on ideas from https://github.com/heketi

import (
	"net/http"
	"net/http/pprof"
	runtime_pprof "runtime/pprof"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	basePath = "/debug/pprof/"
)

func addPath(router *mux.Router, name string, handler http.Handler) {
	router.Path(basePath + name).Name(name).Handler(handler)
	log.WithFields(log.Fields{
		"name":    name,
		"handler": handler,
	}).Debug("Starting golang profiling")
}

//EnableProfiling starts the golang profiling in GD2
func EnableProfiling(router *mux.Router) {
	for _, profile := range runtime_pprof.Profiles() {
		name := profile.Name()
		handler := pprof.Handler(name)
		addPath(router, name, handler)
	}

	addPath(router, "cmdline", http.HandlerFunc(pprof.Cmdline))
	addPath(router, "profile", http.HandlerFunc(pprof.Profile))
	addPath(router, "symbol", http.HandlerFunc(pprof.Symbol))
	addPath(router, "trace", http.HandlerFunc(pprof.Trace))
}
