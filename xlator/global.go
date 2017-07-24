package xlator

// AllOptions contains all possible xlator options for all xlators
// Other packages can directly import this
var AllOptions map[string][]Option

// InitOptions initializes the global variable xlator.AllOptions
func InitOptions() error {
	xopts, err := getAllOptions()
	if err != nil {
		return err
	}
	AllOptions = xopts
	return nil
}
