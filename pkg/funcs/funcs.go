package funcs

func Pointer[T comparable](value T) *T {
	return &value
}
