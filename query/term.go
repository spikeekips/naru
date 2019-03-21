package query

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spikeekips/naru/common"
)

// Hint was derived from native reflect.Kind
type Hint reflect.Kind

const (
	InvalidValue Hint = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
	Time
	Duration
)

func (h Hint) IsValid() bool {
	return h < 1 || h > Duration
}

func (h Hint) String() string {
	s := reflect.Kind(h).String()
	if !strings.HasPrefix(s, "kind") {
		return s
	}

	switch h {
	case Time:
		return "time"
	case Duration:
		return "duration"
	}

	return s
}

type Value struct {
	value interface{}
	hint  Hint
}

func (t Value) MarshalJSON() ([]byte, error) {
	v := t.value
	switch t.hint {
	case Array, Slice:
		var n []interface{}
		for _, i := range t.value.([]Value) {
			b, err := json.Marshal(i)
			if err != nil {
				return nil, err
			}
			n = append(n, json.RawMessage(b))
		}

		v = n
	}

	return json.Marshal(map[string]interface{}{
		"value": v,
		"hint":  t.hint.String(),
	})
}

func NewValue(v interface{}) (Value, error) {
	hint, nv, err := normalizeValue(v)
	if err != nil {
		return Value{}, err
	}

	return Value{value: nv, hint: hint}, nil
}

func (t Value) Value() interface{} {
	return t.value
}

func (t Value) Hint() Hint {
	return t.hint
}

func (t Value) String() string {
	b, _ := json.Marshal(map[string]interface{}{
		"value": t.value,
		"hint":  t.hint.String(),
	})

	return string(b)
}

func (t Value) StringValue() string {
	v, _ := toStringValue(t.hint, t.value)
	return v
}

func (t Value) Equal(v Value) bool {
	return t.value == v.value && t.hint == v.hint
}

type Term struct {
	field string
	value Value
}

func NewTerm(field string, value interface{}) (Term, error) {
	tv, err := NewValue(value)
	if err != nil {
		return Term{}, err
	}

	return Term{field: field, value: tv}, nil
}

func (t Term) Field() string {
	return t.field
}

func (t Term) Value() Value {
	return t.value
}

func (t Term) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"field": t.field,
		"value": map[string]interface{}{
			"value": t.value.Value(),
			"hint":  t.value.Hint().String(),
		},
	})
}

func (t Term) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

func (t Term) Equal(v Term) bool {
	return t.field == v.Field() && t.value.Equal(v.Value())
}

func hintValue(v interface{}) (Hint, error) {
	switch v.(type) {
	case time.Time:
		return Time, nil
	case time.Duration:
		return Duration, nil
	}

	hint := Hint(reflect.ValueOf(v).Kind())

	switch hint {
	case Bool:
	case Int:
	case Int8:
	case Int16:
	case Int32:
	case Int64:
	case Uint:
	case Uint8:
	case Uint16:
	case Uint32:
	case Uint64:
	case Float32:
	case Float64:
	case Complex64:
	case Complex128:
	case String:
	case Array:
	case Slice:
	default:
		return InvalidValue, InvaludValue.New()
	}

	return hint, nil
}

func normalizeValue(v interface{}) (Hint, interface{}, error) {
	hint, err := hintValue(v)
	if err != nil {
		return InvalidValue, nil, err
	}

	switch hint {
	case Int:
	case Int8:
	case Int16:
	case Int32:
	case Int64:
	case Uint:
	case Uint8:
	case Uint16:
	case Uint32:
	case Uint64:
	case Float32:
	case Float64:
	case Complex64:
	case Complex128:
	case Array, Slice:
		var n []Value
		t := reflect.ValueOf(v)
		for i := 0; i < t.Len(); i++ {
			nv, err := NewValue(t.Index(i).Interface())
			if err != nil {
				return InvalidValue, nil, err
			}
			n = append(n, nv)
		}

		return hint, n, nil
	}

	return hint, v, nil
}

func toStringValue(hint Hint, v interface{}) (string, error) {
	switch hint {
	case Bool:
		return strconv.FormatBool(v.(bool)), nil
	case Int:
		return strconv.FormatInt(int64(v.(int)), 10), nil
	case Int8:
		return strconv.FormatInt(int64(v.(int8)), 10), nil
	case Int16:
		return strconv.FormatInt(int64(v.(int16)), 10), nil
	case Int32:
		return strconv.FormatInt(int64(v.(int32)), 10), nil
	case Int64:
		return strconv.FormatInt(int64(v.(int64)), 10), nil
	case Uint:
		return strconv.FormatUint(uint64(v.(uint)), 10), nil
	case Uint8:
		return strconv.FormatUint(uint64(v.(uint8)), 10), nil
	case Uint16:
		return strconv.FormatUint(uint64(v.(uint16)), 10), nil
	case Uint32:
		return strconv.FormatUint(uint64(v.(uint32)), 10), nil
	case Uint64:
		return strconv.FormatUint(uint64(v.(uint64)), 10), nil
	case Float32, Float64:
		return fmt.Sprintf("%f", v), nil
	case Complex64, Complex128:
		return fmt.Sprintf("%v", v), nil
	case Array, Slice:
		b, _ := common.MarshalJSONNotEscapeHTML(v)
		return strings.TrimSpace(string(b)), nil
	case String:
		return v.(string), nil
	case Time:
		return common.FormatISO8601(v.(time.Time)), nil
	case Duration:
		return v.(time.Duration).String(), nil
	}

	return fmt.Sprintf("%v", v), nil
}
