package kslice

func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func Equal[A []T, B []T, T comparable](x []T, y []T) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[T]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}

func Difference[T comparable](slice1 []T, slice2 []T) []T {
	var diff []T

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func Remove[T comparable](slice *[]T, elemsToRemove ...T) {
	for i, elem := range *slice {
		for _, e := range elemsToRemove {
			if e == elem {
				*slice = append((*slice)[:i], (*slice)[i+1:]...)
			}
		}
	}
}

func Contains[T comparable](elems []T, vs ...T) ([]int, bool) {
	idx := []int{}
	for i, s := range elems {
		for _, v := range vs {
			if v == s {
				idx = append(idx, i)
			}
		}
	}
	if len(idx) > 0 {
		return idx, true
	} else {
		return idx, false
	}
}