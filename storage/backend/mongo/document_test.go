package mongostorage

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"testing"

	sebakcommon "boscoin.io/sebak/lib/common"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type testMongoDocument struct {
	suite.Suite
	s *Storage
}

func (t *testMongoDocument) TestMarshal() {
	key := "showme"
	value := map[string]int{
		"1": 1,
		"2": 2,
	}

	doc, err := NewDocument(key, value)
	t.NoError(err)

	{
		b, err := bson.Marshal(doc)
		t.NoError(err)
		t.NotEmpty(b)
	}

	{
		b, err := bson.MarshalExtJSON(doc, true, true)
		t.NoError(err)
		t.NotEmpty(b)
	}
}

func (t *testMongoDocument) TestUnmarshal() {
	key := "showme"
	value := map[string]int{
		"1": 1,
		"2": 2,
	}

	doc, err := NewDocument(key, value)
	t.NoError(err)

	{
		b, err := bson.Marshal(doc)
		t.NoError(err)
		t.NotEmpty(b)

		var doc Document
		err = bson.Unmarshal(b, &doc)
		t.NoError(err)
		t.Equal(key, doc.Key())

		var returned map[string]int
		unmarshaledDoc, err := UnmarshalDocument(b, &returned)
		t.NoError(err)
		t.Equal(key, unmarshaledDoc.Key())
		t.Equal(value, unmarshaledDoc.Value())
		t.Equal(value, returned)

		for k, v := range unmarshaledDoc.Value().(map[string]int) {
			t.Equal(value[k], v)
		}
		for k, v := range value {
			t.Equal(returned[k], v)
		}
	}
}

type testUnmarshalStruct struct {
	A string
	B int
	C []uint64
}

func (t *testMongoDocument) TestUnmarshalStruct() {
	key := "showme"
	value := testUnmarshalStruct{
		A: "AAA",
		B: 99,
		C: []uint64{7, 8, 9},
	}

	doc, err := NewDocument(key, value)
	t.NoError(err)

	{
		b, err := bson.Marshal(doc)
		t.NoError(err)
		t.NotEmpty(b)

		var doc *Document
		err = bson.Unmarshal(b, &doc)
		t.NoError(err)
		t.Equal(key, doc.Key())

		{
			var returned testUnmarshalStruct
			unmarshaledDoc, err := UnmarshalDocument(b, &returned)
			t.NoError(err)
			t.Equal(key, unmarshaledDoc.Key())
			t.Equal(value, unmarshaledDoc.Value())
			t.Equal(value, returned)

			t.Equal(value.A, returned.A)
			t.Equal(value.B, returned.B)
			t.Equal(value.C, returned.C)
		}
	}
}

func TestMongoDocument(t *testing.T) {
	suite.Run(t, new(testMongoDocument))
}

type testMongoValue struct {
	suite.Suite
}

func (t *testMongoValue) TestEncodeDecodeValue() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r)
		}
	}()

	cases := []struct {
		input interface{}
		err   error
		t     Hint
		m     bsontype.Type
		msg   string
	}{
		{
			input: true,
			err:   nil,
			t:     Bool,
			m:     bsontype.Boolean,
			msg:   "bool",
		},
		{
			input: int(10),
			err:   nil,
			t:     Int,
			m:     bsontype.Int32,
			msg:   "int",
		},
		{
			input: int8(10),
			err:   nil,
			t:     Int8,
			m:     bsontype.Int32,
			msg:   "int8",
		},
		{
			input: int16(10),
			err:   nil,
			t:     Int16,
			m:     bsontype.Int32,
			msg:   "int16",
		},
		{
			input: int32(10),
			err:   nil,
			t:     Int32,
			m:     bsontype.Int32,
			msg:   "int32",
		},
		{
			input: int64(10),
			err:   nil,
			t:     Int64,
			m:     bsontype.Int64,
			msg:   "int64",
		},
		{
			input: uint(10),
			err:   nil,
			t:     Uint,
			m:     bsontype.Decimal128,
			msg:   "uint",
		},
		{
			input: uint8(10),
			err:   nil,
			t:     Uint8,
			m:     bsontype.Decimal128,
			msg:   "uint8",
		},
		{
			input: uint16(10),
			err:   nil,
			t:     Uint16,
			m:     bsontype.Decimal128,
			msg:   "uint16",
		},
		{
			input: uint32(10),
			err:   nil,
			t:     Uint32,
			m:     bsontype.Decimal128,
			msg:   "uint32",
		},
		{
			input: uint64(10000000000000000000),
			err:   nil,
			t:     Uint64,
			m:     bsontype.Decimal128,
			msg:   "uint64",
		},
		{
			input: float32(10.01),
			err:   nil,
			t:     Float32,
			m:     bsontype.Double,
			msg:   "float32",
		},
		{
			input: float64(1000.02),
			err:   nil,
			t:     Float64,
			m:     bsontype.Double,
			msg:   "float64",
		},
		{
			input: map[string]int{"a": 99},
			err:   nil,
			t:     Map,
			m:     bsontype.EmbeddedDocument,
			msg:   "map",
		},
		{
			input: "killme",
			err:   nil,
			t:     String,
			m:     bsontype.String,
			msg:   "string",
		},
		{
			input: struct {
				A string `bson:"a"`
				B []int  `bson:"b"`
			}{A: "showme", B: []int{1, 2}},
			err: nil,
			t:   Struct,
			m:   bsontype.EmbeddedDocument,
			msg: "struct: showme",
		},
		{
			input: []int{1, 2},
			err:   nil,
			t:     Slice,
			m:     bsontype.Array,
			msg:   "array",
		},
		{
			input: make(chan bool),
			err:   nil,
			t:     Chan,
			m:     bsontype.Undefined,
			msg:   "make(chan bool)",
		},
		{
			input: &Value{},
			err:   nil,
			t:     Ptr,
			m:     bsontype.EmbeddedDocument,
			msg:   "&Value{}",
		},
	}

	for n, c := range cases {
		bt, encoded, err := encodeValue(c.input)
		if c.err == nil {
			t.NoError(err, c.msg)
		} else {
			t.EqualError(c.err, err.Error(), c.msg)
		}

		t.Equal(c.m, bt, c.msg)

		if c.err != nil || c.m == bsontype.Undefined {
			continue
		}

		b, err := bson.Marshal(bson.M{"a": "showme", "b": encoded})
		t.NoError(err, "%s: %d", c.msg, n)

		raw := bson.Raw(b)
		rv := raw.Lookup("b")

		f := reflect.New(reflect.TypeOf(c.input)).Interface()
		err = decodeValue(rv, f)
		t.NoError(err, "%s: %d", c.msg, n)

		e := reflect.ValueOf(f).Elem().Interface()
		t.Equal(c.input, e, c.msg, strconv.Itoa(n))
		t.Equal(c.t, Hint(reflect.TypeOf(e).Kind()), "%s: %d", c.msg, n)
	}
}

func (t *testMongoValue) TestMarshalStruct() {
	input := struct {
		A string `bson:"a"`
		B []int  `bson:"b"`
	}{A: "showme", B: []int{1, 2}}

	b, err := bson.Marshal(input)
	t.NoError(err)

	{
		f := reflect.New(reflect.TypeOf(input)).Elem().Interface()
		err := bson.Unmarshal(b, &f)
		t.NoError(err)
		t.Equal(reflect.Slice, reflect.TypeOf(f).Kind())
	}

	{
		f := reflect.New(reflect.TypeOf(input)).Interface()
		err := bson.Unmarshal(b, f)
		t.NoError(err)
		t.Equal(reflect.Ptr, reflect.TypeOf(f).Kind())
		t.Equal(reflect.Struct, reflect.ValueOf(f).Elem().Kind())
	}

	{
		type inputStruct struct {
			A string `bson:"a"`
			B []int  `bson:"b"`
		}

		f := reflect.New(reflect.TypeOf(inputStruct{})).Elem().Interface()
		err := bson.Unmarshal(b, &f)
		t.NoError(err)
		t.Equal(reflect.Slice, reflect.TypeOf(f).Kind())
	}
}

/*

func (a Amount) MarshalBSONValue() (bsontype.Type, []byte, error) {
	n, err := primitive.ParseDecimal128(a.String())
	if err != nil {
		return bsontype.Decimal128, nil, err
	}

	return bsontype.Decimal128, bsoncore.AppendDecimal128(nil, n), nil
}

func (a *Amount) UnmarshalBSON(b []byte) error {
	val := bson.RawValue{
		Type:  bsontype.Decimal128,
		Value: b,
	}
	var d primitive.Decimal128
	if err := val.Unmarshal(&d); err != nil {
		return err
	}

	c, err := strconv.ParseUint(d.String(), 10, 64)
	if err != nil {
		return err
	}

	*a = Amount(c)

	return nil
}
*/

func (t *testMongoValue) TestEncodeDecodeSEBAKAmount() {
	{ // with using bson default registry
		rb := bson.NewRegistryBuilder()
		registry := rb.Build()

		var k uint64 = 10000000000000000000
		amount := sebakcommon.Amount(k)
		b, err := bson.MarshalWithRegistry(registry, bson.M{"A": &amount})
		t.Error(err, "10000000000000000000 overflows int64")
		t.Empty(b)
	}

	{ // with custom encoder
		var tSEBAKAmount reflect.Type = reflect.TypeOf(sebakcommon.Amount(0))
		SEBAKAmountEncodeValue := func(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
			if !val.IsValid() || val.Type() != tSEBAKAmount {
				return bsoncodec.ValueEncoderError{Name: "SEBAKAmountEncodeValue", Types: []reflect.Type{tSEBAKAmount}, Received: val}
			}

			n, err := primitive.ParseDecimal128(val.Interface().(sebakcommon.Amount).String())
			if err != nil {
				return err
			}
			return vw.WriteDecimal128(n)
		}

		rb := bson.NewRegistryBuilder()
		rb.RegisterEncoder(tSEBAKAmount, bsoncodec.ValueEncoderFunc(SEBAKAmountEncodeValue))
		registry := rb.Build()

		var k uint64 = 10000000000000000000
		amount := sebakcommon.Amount(k)
		{
			b, err := bson.MarshalWithRegistry(registry, bson.M{"A": &amount})
			t.NoError(err)
			t.NotEmpty(b)
		}
		{
			b, err := bson.MarshalWithRegistry(registry, bson.M{"A": amount})
			t.NoError(err)
			t.NotEmpty(b)
		}
	}

	{ // with custom decoder
		var tSEBAKAmount reflect.Type = reflect.TypeOf(sebakcommon.Amount(0))
		SEBAKAmountEncodeValue := func(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
			if !val.IsValid() || val.Type() != tSEBAKAmount {
				return bsoncodec.ValueEncoderError{Name: "SEBAKAmountEncodeValue", Types: []reflect.Type{tSEBAKAmount}, Received: val}
			}

			n, err := primitive.ParseDecimal128(val.Interface().(sebakcommon.Amount).String())
			if err != nil {
				return err
			}
			return vw.WriteDecimal128(n)
		}

		SEBAKAmountDecodeValue := func(dctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
			if vr.Type() != bsontype.Decimal128 {
				return fmt.Errorf("cannot decode %v into a primitive.Decimal128", vr.Type())
			}

			if !val.CanSet() || val.Type() != tSEBAKAmount {
				return bsoncodec.ValueDecoderError{Name: "SEBAKAmountDecodeValue", Types: []reflect.Type{tSEBAKAmount}, Received: val}
			}

			d128, err := vr.ReadDecimal128()

			c, err := strconv.ParseUint(d128.String(), 10, 64)
			if err != nil {
				return err
			}

			val.Set(reflect.ValueOf(sebakcommon.Amount(c)))
			return nil
		}

		rb := bson.NewRegistryBuilder()
		rb.RegisterEncoder(tSEBAKAmount, bsoncodec.ValueEncoderFunc(SEBAKAmountEncodeValue))
		rb.RegisterDecoder(tSEBAKAmount, bsoncodec.ValueDecoderFunc(SEBAKAmountDecodeValue))
		registry := rb.Build()

		var k uint64 = 10000000000000000000
		amount := sebakcommon.Amount(k)
		encoded, _ := bson.MarshalWithRegistry(
			registry,
			struct {
				A sebakcommon.Amount
			}{A: amount})

		decodedAmount := &(struct {
			A sebakcommon.Amount
		}{})

		err := bson.UnmarshalWithRegistry(registry, encoded, decodedAmount)
		t.NoError(err)
		t.Equal(amount, decodedAmount.A)
	}
}

func (t *testMongoValue) TestSerializeSEBAKAmount() {
	var k uint64 = 10000000000000000000
	amount := sebakcommon.Amount(k)
	var encoded []byte
	{
		var err error
		encoded, err = Serialize(
			struct {
				A sebakcommon.Amount
			}{A: amount})
		t.NoError(err)
		t.NotEmpty(encoded)
	}

	{
		decodedAmount := &(struct {
			A sebakcommon.Amount
		}{})

		err := Deserialize(encoded, decodedAmount)
		t.NoError(err)
		t.Equal(amount, decodedAmount.A)
	}

}

func TestMongoValue(t *testing.T) {
	suite.Run(t, new(testMongoValue))
}
