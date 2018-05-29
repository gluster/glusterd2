package daemon

// storedDaemon is used to save/retrieve a daemons information in the store,
// and also implements the Daemon interface
type storedDaemon struct {
	DName, DPath, DSocketFile, DPidFile, DID string

	DArgs []string
}

func newStoredDaemon(d Daemon) *storedDaemon {
	return &storedDaemon{
		DName:       d.Name(),
		DPath:       d.Path(),
		DArgs:       d.Args(),
		DSocketFile: d.SocketFile(),
		DPidFile:    d.PidFile(),
		DID:         d.ID(),
	}
}

func (s *storedDaemon) Name() string {
	return s.DName
}

func (s *storedDaemon) Path() string {
	return s.DPath
}

func (s *storedDaemon) Args() []string {
	return s.DArgs
}

func (s *storedDaemon) SocketFile() string {
	return s.DSocketFile
}

func (s *storedDaemon) PidFile() string {
	return s.DPidFile
}

func (s *storedDaemon) ID() string {
	return s.DID
}
