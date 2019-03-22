package mongostorage

import (
	"errors"
	"reflect"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type Document struct {
	K string      `bson:"_k"`
	V interface{} `bson:"_v"`
}

func NewDocument(key string, value interface{}) (*Document, error) {
	_, encoded, err := encodeValue(value)
	if err != nil {
		return nil, err
	}

	if _, err := Serialize(bson.M{"_v": encoded}); err != nil {
		return nil, err
	}

	return &Document{K: key, V: value}, nil
}

func (d *Document) Key() string {
	return d.K
}

func (d *Document) Value() interface{} {
	return d.V
}

func (d *Document) BSONDocument() bson.M {
	_, encoded, _ := encodeValue(d.V)
	return bson.M{"_k": d.K, "_v": encoded}
}

func (d *Document) MarshalBSON() ([]byte, error) {
	return Serialize(d.BSONDocument())
}

func UnmarshalDocument(b []byte, v interface{}) (*Document, error) {
	key, fv, err := validateDocumentByRaw(b)
	if err != nil {
		return nil, err
	}

	if err := decodeValue(fv, v); err != nil {
		return nil, err
	}

	return &Document{K: key, V: reflect.ValueOf(v).Elem().Interface()}, nil
}

func UnmarshalDocumentValue(b []byte, v interface{}) (string, error) {
	key, fv, err := validateDocumentByRaw(b)
	if err != nil {
		return "", err
	}

	if err := decodeValue(fv, v); err != nil {
		return "", err
	}

	return key, nil
}

func validateDocumentByRaw(b []byte) (string, bson.RawValue, error) {
	raw := bson.Raw(b)
	if err := raw.Validate(); err != nil {
		return "", bson.RawValue{}, err
	}

	var key string

	fk, err := raw.LookupErr("_k")
	if err != nil {
		return "", bson.RawValue{}, err
	} else if k, ok := fk.StringValueOK(); !ok {
		return "", bson.RawValue{}, InvalidDocumentKey.New()
	} else {
		key = k
	}

	fv, err := raw.LookupErr("_v")
	if err != nil {
		return "", bson.RawValue{}, err
	}

	return key, fv, nil
}

type Value struct {
	I  interface{}
	M  interface{}
	T  Hint
	BT bsontype.Type
}

func encodeValue(v interface{}) (bsontype.Type, interface{}, error) {
	var err error
	var m interface{} = v
	var bt bsontype.Type

	switch Hint(reflect.ValueOf(v).Kind()) {
	case Bool:
		bt = bsontype.Boolean
	case Int:
		bt = bsontype.Int32
	case Int8:
		bt = bsontype.Int32
	case Int16:
		bt = bsontype.Int32
	case Int32:
		bt = bsontype.Int32
	case Int64:
		bt = bsontype.Int64
	case Uint:
		bt = bsontype.Decimal128
		m, err = primitive.ParseDecimal128(strconv.FormatUint(uint64(v.(uint)), 10))
	case Uint8:
		bt = bsontype.Decimal128
		m, err = primitive.ParseDecimal128(strconv.FormatUint(uint64(v.(uint8)), 10))
	case Uint16:
		bt = bsontype.Decimal128
		m, err = primitive.ParseDecimal128(strconv.FormatUint(uint64(v.(uint16)), 10))
	case Uint32:
		bt = bsontype.Decimal128
		m, err = primitive.ParseDecimal128(strconv.FormatUint(uint64(v.(uint32)), 10))
	case Uint64:
		bt = bsontype.Decimal128
		m, err = primitive.ParseDecimal128(strconv.FormatUint(uint64(v.(uint64)), 10))
	//case reflect.Uintptr:
	case Float32, Float64:
		bt = bsontype.Double
	//case reflect.Complex64, reflect.Complex128:
	//case reflect.Chan:
	//case reflect.Func:
	//case reflect.Interface:
	case Map:
		bt = bsontype.EmbeddedDocument
	case Ptr:
		return encodeValue(reflect.ValueOf(v).Elem().Interface())
	case String:
		bt = bsontype.String
	case Struct:
		bt = bsontype.EmbeddedDocument
	//case reflect.UnsafePointer:
	case Array, Slice:
		bt = bsontype.Array
	default:
		bt = bsontype.Undefined
	}

	if err != nil {
		return bsontype.Undefined, nil, err
	}

	return bt, m, nil
}

func convertDecimal128ToUint(rv bson.RawValue, bitSize int) (interface{}, error) {
	d, ok := rv.Decimal128OK()
	if !ok {
		return nil, errors.New("not primitive.Decimal128 type")
	}
	return strconv.ParseUint(d.String(), 10, bitSize)
}

func decodeValue(rv bson.RawValue, f interface{}) error {
	var o interface{}

	// TODO support mongo timestamp

	switch Hint(reflect.TypeOf(f).Elem().Kind()) {
	case Uint:
		if i, err := convertDecimal128ToUint(rv, 32); err != nil {
			return err
		} else {
			o = uint(i.(uint64))
		}
	case Uint8:
		if i, err := convertDecimal128ToUint(rv, 8); err != nil {
			return err
		} else {
			o = uint8(i.(uint64))
		}
	case Uint16:
		if i, err := convertDecimal128ToUint(rv, 16); err != nil {
			return err
		} else {
			o = uint16(i.(uint64))
		}
	case Uint32:
		if i, err := convertDecimal128ToUint(rv, 32); err != nil {
			return err
		} else {
			o = uint32(i.(uint64))
		}
	case Uint64:
		if i, err := convertDecimal128ToUint(rv, 64); err != nil {
			return err
		} else {
			o = i.(uint64)
		}
	default:
		return rv.UnmarshalWithRegistry(DefaultBSONRegistry, f)
	}

	reflect.ValueOf(f).Elem().Set(reflect.ValueOf(o))

	return nil
}
