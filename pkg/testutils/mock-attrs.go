package testutils

// MockRemovexattr is mock function for unix.Removexattr
func MockRemovexattr(path string, attr string) (err error) {
	return nil
}

// MockSetxattr is mock function for unix.Setxattr
func MockSetxattr(path string, attr string, data []byte, flags int) (err error) {
	return nil
}

// MockGetxattr is mock function for unix.Getxattr
func MockGetxattr(path string, attr string, dest []byte) (sz int, err error) {
	return 0, nil
}

// MockValidateBrickPathStats is mock function for utils.ValidateBrickPathStats
func MockValidateBrickPathStats(brickPath string, host string, force bool) error {
	return nil
}
