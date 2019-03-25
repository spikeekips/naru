package storage

type Record struct {
	Key   string
	Value interface{}
}

func NewRecord(key string, value interface{}) Record {
	return Record{Key: key, Value: value}
}

type Value struct {
	Key   string
	Value interface{}
}

func NewValue(key string, value interface{}) Value {
	return Value{Key: key, Value: value}
}
