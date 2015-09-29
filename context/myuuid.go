package context

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/gluster/glusterd2/config"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

var (
	myUUIDFile = path.Join(config.LocalStateDir, "uuid")
)

// initMyUUID initializes MyUUID by reading the `<config.LocalStateDir>/uuid` file.
// If the file is not present it generates a new UUID and saves it to the file.
// If the file is not accessible. initMyUUID panics.
func initMyUUID() {
	ubytes, err := ioutil.ReadFile(myUUIDFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			genMyUUID()
			return

		default:
			log.WithFields(log.Fields{
				"err":  err,
				"path": myUUIDFile,
			}).Fatal("failed to read MyUUID from file")
		}
	}

	MyUUID = uuid.Parse(string(ubytes))
	log.WithField("myuuid", MyUUID.String()).Info("restored MyUUID")
}

func genMyUUID() {
	MyUUID = uuid.NewRandom()
	writeMyUUIDFile()
	log.WithField("myuuid", MyUUID.String()).Info("generated new MyUUID")
}

func writeMyUUIDFile() {
	if err := ioutil.WriteFile(myUUIDFile, []byte(MyUUID.String()), os.ModePerm); err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"path": myUUIDFile,
		}).Fatal("failed to write MyUUID to file")
	}
	log.WithFields(log.Fields{
		"myuuid": MyUUID.String(),
		"path":   myUUIDFile,
	}).Debug("wrote MyUUID to file")
}
