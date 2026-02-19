package cache

type Stub[T any] struct{}

func (c *Stub[T]) Add(_ string, _ T) {
}

func (c *Stub[T]) Get(_ string) (T, bool) { //nolint:ireturn
	var zero T
	return zero, false
}

func (c *Stub[T]) Del(_ string) {
}

func (c *Stub[T]) Cleanup() {
}

func (c *Stub[T]) Size() int {
	return 0
}

func NewStub[T any]() *Stub[T] {
	return &Stub[T]{}
}
