package utils

// MergeStringMaps merges give string maps and returns a single string map
// If maps have same keys, the merged map will have the value of the last map the key was present in.
func MergeStringMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}
