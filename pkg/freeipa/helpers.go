package freeipa

import (
	"fmt"
	"reflect"
	"slices"
)

func isBool(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Bool
}

func isNotEmptySlice(v any) bool {
	if reflect.TypeOf(v).Kind() == reflect.Slice {
		if sl, ok := v.([]any); ok {
			return len(sl) > 0
		}
	}
	return false
}

func convertSliceAnyToSliceStr(vSrc []any) []string {
	result := make([]string, len(vSrc))
	for i, v := range vSrc {
		result[i] = fmt.Sprint(v)
	}
	return result
}

func getRangeFromSlice[T any](s []T, limitSrc, offsetSrc, defaultLimit int32) []T {
	limit := defaultLimit
	var offset int32 = 0
	newS := slices.Clone(s)

	if limitSrc > 0 {
		limit = limitSrc
	}
	if offsetSrc > 0 {
		offset = offsetSrc
	}

	var result []T

	for i, v := range newS {
		i32 := int32(i)
		if i32 >= offset && i32 < offset+limit {
			result = append(result, v)
		}
	}

	return result
}
