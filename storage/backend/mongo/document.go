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

	if _, err := bson.Marshal(bson.M{"_v": encoded}); err != nil {
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
	return bson.Marshal(d.BSONDocument())
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

/*
func validateDocumentByRaw(b []byte) (string, bson.Raw, error) {
	raw := bson.Raw(b)
	if err := raw.Validate(); err != nil {
		return "", bson.Raw{}, err
	}

	var key string

	fk, err := raw.LookupErr("_k")
	if err != nil {
		return "", bson.Raw{}, err
	} else if k, ok := fk.StringValueOK(); !ok {
		return "", bson.Raw{}, InvalidDocumentKey.New()
	} else {
		key = k
	}

	var v bson.Raw
	fv, err := raw.LookupErr("_v")
	if err != nil {
		return "", bson.Raw{}, err
	} else if d, ok := fv.DocumentOK(); !ok {
		fmt.Println("ddddddddddddeeeeeeeee", string(fv.Value))
		return "", bson.Raw{}, InvalidDocumentValue.New()
	} else {
		v = d
	}

	return key, v, nil
}
*/

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

/*
func DecodeMongoValue(m f interface{}) error {
	var err error
	var o interface{} = v.I

	switch v.T {
	case Bool:
		o = v.I.(bool)
	case Int:
		o = int(v.I.(int32))
	case Int8:
		o = int8(v.I.(int32))
	case Int16:
		o = int16(v.I.(int32))
	case Int32:
		o = v.I.(int32)
	case Int64:
		o = v.I.(int64)
	case Uint:
		if i, err := convertDecimal128ToUint(v.I, 32); err != nil {
			return err
		} else {
			o = uint(i.(uint64))
		}
	case Uint8:
		if i, err := convertDecimal128ToUint(v.I, 8); err != nil {
			return err
		} else {
			o = uint8(i.(uint64))
		}
	case Uint16:
		if i, err := convertDecimal128ToUint(v.I, 16); err != nil {
			return err
		} else {
			o = uint16(i.(uint64))
		}
	case Uint32:
		if i, err := convertDecimal128ToUint(v.I, 32); err != nil {
			return err
		} else {
			o = uint32(i.(uint64))
		}
	case Uint64:
		if i, err := convertDecimal128ToUint(v.I, 64); err != nil {
			return err
		} else {
			o = i.(uint64)
		}
		//case reflect.Uintptr:
	case Float32:
		switch v.I.(type) {
		case float32:
		case float64:
			o = float32(v.I.(float64))
		}
	case Float64:
		o = v.I.(float64)
	//case reflect.Complex64, reflect.Complex128:
	//case reflect.Chan:
	//case reflect.Func:
	//case reflect.Interface:
	case Map:
		var b []byte
		if b, err = bson.Marshal(v.I); err != nil {
			return err
		}

		if err = bson.Unmarshal(b, f); err != nil {
			return err
		}
		return nil
	case Ptr:
	case String:
	case Struct:
		var b []byte
		if b, err = bson.Marshal(v.I); err != nil {
			return err
		}

		if err = bson.Unmarshal(b, f); err != nil {
			return err
		}
		return nil
	//case reflect.UnsafePointer:
	case Array, Slice:
		var b []byte
		if b, err = bson.Marshal(bson.M{"a": v.I}); err != nil {
			return err
		}

		if err = bson.Raw(b).Lookup("a").Unmarshal(f); err != nil {
			return err
		}
		return nil
	default:
		return InvalidDocumentValue.New()
	}

	reflect.ValueOf(f).Elem().Set(reflect.ValueOf(o))

	return nil
}
*/

func convertDecimal128ToUint0(i interface{}, bitSize int) (interface{}, error) {
	if d, ok := i.(primitive.Decimal128); !ok {
		return nil, errors.New("not primitive.Decimal128 type")
	} else {
		return strconv.ParseUint(d.String(), 10, bitSize)
	}
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
		return rv.Unmarshal(f)
	}

	reflect.ValueOf(f).Elem().Set(reflect.ValueOf(o))

	return nil
}
