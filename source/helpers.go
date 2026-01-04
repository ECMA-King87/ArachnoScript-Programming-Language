package main

import "slices"

func is_value[T comparable](value T, values ...T) bool {
	bool := slices.Contains(values, value)
	return bool
}
