package utils

import "reflect"

// GetTypeString returns the type of instance passed, as a string.
// Go doesn't have type literals. Hence one has to pass (*Type)(nil)
// as argument to this function.
func GetTypeString(i interface{}) string {
	return reflect.TypeOf(i).Elem().String()
}
