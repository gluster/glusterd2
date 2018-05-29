package api

import "strings"

// mapSize calculates and returns the size of the map passed as parameter in bytes
func mapSize(metadata map[string]string) int {
	size := 0
	for key, value := range metadata {
		if !strings.HasPrefix(key, "_") {
			size = size + len(key) + len(value)
		}
	}
	return size
}
