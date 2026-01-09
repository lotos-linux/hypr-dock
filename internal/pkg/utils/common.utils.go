package utils

import "slices"

func AddToSlice(slice *[]string, value string) {
	*slice = append(*slice, value)
}

func RemoveFromSliceByValue(slice *[]string, value string) {
	index := -1
	for i, v := range *slice {
		if v == value {
			index = i
			break
		}
	}

	if index != -1 {
		*slice = append((*slice)[:index], (*slice)[index+1:]...)
	}
}

func RemoveFromSliceByFunc[T any](slice *[]T, shouldRemove func(T) bool) {
	for i, v := range *slice {
		if shouldRemove(v) {
			*slice = slices.Delete(*slice, i, i+1)
			return
		}
	}
}

func RemoveFromSlice(slice []map[string]string, s int) []map[string]string {
	return append(slice[:s], slice[s+1:]...)
}

func GetSingleValue[K comparable, V any](m map[K]V) (V, bool) {
	for _, v := range m {
		return v, true
	}
	var zero V
	return zero, false
}
