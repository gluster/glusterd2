package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeStringMaps(t *testing.T) {
	map1 := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}
	map2 := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
		"d": "d",
	}

	mergedmap := MergeStringMaps(map1, map2)
	assert.Equal(t, mergedmap, map1)

	map2 = map[string]string{
		"e": "e",
		"f": "f",
		"g": "g",
		"h": "h",
	}

	mergedmap = MergeStringMaps(map1, map2)
	assert.Equal(t, len(mergedmap), 8)
}
