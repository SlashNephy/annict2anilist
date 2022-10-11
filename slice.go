package main

import "golang.org/x/exp/slices"

func CountByKey[T any](slice []T, key func(T) bool) int {
	count := 0
	for _, t := range slice {
		if key(t) {
			count++
		}
	}

	return count
}

func Contains[T any](slice []T, key func(T) bool) bool {
	return slices.IndexFunc(slice, key) >= 0
}

func FindByKey[T any](slice []T, key func(T) bool) (*T, bool) {
	index := slices.IndexFunc(slice, key)
	if index < 0 {
		return nil, false
	}

	return &slice[index], true
}
