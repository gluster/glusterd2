package utils

import (
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// WriteStatedump writes statedump information to a file. The file name is
// of the format glusterd2.<pid>.dump.<timestamp>. This file will be
// written to the directory passed.
func WriteStatedump(dirpath string) {

	// Run the expvar http handler
	w := httptest.NewRecorder()
	expvar.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/statedump", nil))
	respBody, err := ioutil.ReadAll(w.Result().Body)
	if err != nil {
		log.WithError(err).Error("Failed to fetch statedump details from expvar handler")
		return
	}

	dumpFileName := fmt.Sprintf("glusterd2.%s.dump.%s",
		strconv.Itoa(os.Getpid()), strconv.Itoa(int(time.Now().Unix())))
	dumpPath := path.Join(dirpath, dumpFileName)

	if err := ioutil.WriteFile(dumpPath, respBody, 0644); err != nil {
		log.WithError(err).WithField("file", dumpPath).Error("Failed to write statedump to file")
		return
	}
	log.WithField("file", dumpPath).Info("Statedump written to file")
}
