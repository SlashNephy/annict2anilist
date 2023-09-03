package main

import "slices"

func CountByKey[T any](slice []T, key func(T) bool) int {
	count := 0
	for _, t := range slice {
		if key(t) {
			count++
		}
	}

	return count
}

func FindByKey[T any](slice []T, key func(T) bool) (*T, bool) {
	index := slices.IndexFunc(slice, key)
	if index < 0 {
		return nil, false
	}

	return &slice[index], true
}
