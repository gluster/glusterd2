package georeplication

// ErrGeorepSessionNotFound custom error to represent georep session not exists
// in store
type ErrGeorepSessionNotFound struct{}

func (e *ErrGeorepSessionNotFound) Error() string {
	return "geo-replication session not found"
}
