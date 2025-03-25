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

// MapFn applies a function to each element of a slice and returns a new slice.
func MapFn[T any, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}
