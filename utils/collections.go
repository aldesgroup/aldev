// ----------------------------------------------------------------------------
// General utilities for handling collections
// ----------------------------------------------------------------------------
package utils

// InSlice returns true if the slice s contains the given element el
func InSlice[V comparable](s []V, el V) bool {
	for _, val := range s {
		if val == el {
			return true
		}
	}

	return false
}
