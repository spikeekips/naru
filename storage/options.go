package storage

var DefaultMaxLimitListOptions uint64 = 100

type ListOptions interface {
	Reverse() bool
	SetReverse(bool) ListOptions
	Cursor() []byte
	SetCursor([]byte) ListOptions
	Limit() uint64
	SetLimit(uint64) ListOptions
}

type DefaultListOptions struct {
	reverse bool
	cursor  []byte
	limit   uint64
}

func NewDefaultListOptions(reverse bool, cursor []byte, limit uint64) *DefaultListOptions {
	return &DefaultListOptions{
		reverse: reverse,
		cursor:  cursor,
		limit:   limit,
	}
}

func (o DefaultListOptions) Reverse() bool {
	return o.reverse
}

func (o *DefaultListOptions) SetReverse(r bool) ListOptions {
	o.reverse = r
	return o
}

func (o DefaultListOptions) Cursor() []byte {
	return o.cursor
}

func (o *DefaultListOptions) SetCursor(c []byte) ListOptions {
	o.cursor = c
	return o
}

func (o DefaultListOptions) Limit() uint64 {
	return o.limit
}

func (o *DefaultListOptions) SetLimit(l uint64) ListOptions {
	o.limit = l
	return o
}
