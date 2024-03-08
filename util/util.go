package util

func Some[T any](s []T, f func(T) bool) bool {
	for _, e := range s {
		if f(e) {
			return true
		}
	}
	return false
}
