package elasticetcd

import (
	"testing"
)

func TestCompareStringSlices(t *testing.T) {
	a := []string{"1", "2", "3", "4"}
	b := []string{"5", "6", "7", "8"}

	tests := []struct {
		a, b     []string
		expected bool
	}{
		{a, a, true},
		{a, b, false},
		{a, a[0:2], false},
	}

	for _, i := range tests {
		r := compareStringSlices(i.a, i.b)
		if r != i.expected {
			t.Errorf("compareStringSlices(%v, %v): expected %v, got %v", i.a, i.b, i.expected, r)
		}
	}
}

func TestDiffStringSlices(t *testing.T) {
	a := []string{"1", "2", "3", "4"}
	b := []string{"5", "6", "7", "8"}

	tests := []struct {
		a, b, expected []string
	}{
		{a, a, []string{}},
		{a, b, a},
		{a, a[0:2], []string{"3", "4"}},
		{a, append(a, b...), []string{}},
		{[]string{"3"}, a, []string{}},
	}

	for _, i := range tests {
		r := diffStringSlices(i.a, i.b)
		if !compareStringSlices(i.expected, r) {
			t.Errorf("diffStringSlices(%v, %v): expected %v, got %v", i.a, i.b, i.expected, r)
		}
	}
}
