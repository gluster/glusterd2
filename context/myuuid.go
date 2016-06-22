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

// InitMyUUID initializes MyUUID by reading the `<config.LocalStateDir>/uuid` file.
// If the file is not present it generates a new UUID and saves it to the file.
// If the file is not accessible. initMyUUID panics.
func InitMyUUID() uuid.UUID {
	ubytes, err := ioutil.ReadFile(myUUIDFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			u := genMyUUID()
			return u

		default:
			log.WithFields(log.Fields{
				"err":  err,
				"path": myUUIDFile,
			}).Fatal("failed to read MyUUID from file")
		}
	}

	u := uuid.Parse(string(ubytes))
	log.WithField("myuuid", u.String()).Info("restored uuid")
	Restart = true

	return u
}

func genMyUUID() uuid.UUID {
	u := uuid.NewRandom()

	writeMyUUIDFile(u)
	log.WithField("myuuid", u.String()).Info("generated new MyUUID")
	return u
}

func writeMyUUIDFile(u uuid.UUID) {
	if err := ioutil.WriteFile(myUUIDFile, []byte(u.String()), 0644); err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"path": myUUIDFile,
		}).Fatal("failed to write MyUUID to file")
	}
	log.WithFields(log.Fields{
		"myuuid": u.String(),
		"path":   myUUIDFile,
	}).Debug("wrote MyUUID to file")
}
