package context

import (
	"io/ioutil"
	"os"
	"syscall"

	"github.com/gluster/glusterd2/config"

	log "github.com/Sirupsen/logrus"
)

// initLocalStateDir checks if `config.LocalStateDir` is present, a direcotry and is accessible.
// If the directory is not present, it is created.
// If it is not a directory, initLocalStateDir panics.
// If the directory is not accessible, initLocalStateDir panics.
func initLocalStateDir() {
	di, err := os.Stat(config.LocalStateDir)

	if err != nil {
		switch {
		case os.IsNotExist(err):
			if err = os.Mkdir(config.LocalStateDir, os.ModeDir|os.ModePerm); err != nil {
				log.WithFields(log.Fields{
					"err":  err,
					"path": config.LocalStateDir,
				}).Fatal("failed to create local state directory")
			}
			return

		case os.IsPermission(err):
			log.WithFields(log.Fields{
				"err":  err,
				"path": config.LocalStateDir,
			}).Fatal("failed to access local state directory")
		}
	}

	if !di.IsDir() {
		log.WithFields(log.Fields{
			"err":  syscall.ENOTDIR,
			"path": config.LocalStateDir,
		}).Fatal("local state directory path is not a directory")
	}

	// Check if you can create entries in `config.LocalStateDir`
	t, err := ioutil.TempFile(config.LocalStateDir, "")
	if err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"path": config.LocalStateDir,
		}).Fatal("local state directory path is not a writable")
	}
	// defer happens in LIFO
	defer syscall.Unlink(t.Name())
	defer t.Close()
}
