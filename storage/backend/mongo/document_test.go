package mongostorage

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"testing"

	sebakcommon "boscoin.io/sebak/lib/common"
	"github.com/spikeekips/naru/common"
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

func (t *testMongoDocument) TestMarshalWithBSONMarshaller() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r)
		}
	}()

	key := "showme"
	value := map[string]uint64{
		"1": 1,
		"2": 2,
	}

	_, err := Serialize(value)
	t.NoError(err)

	origDoc, err := NewDocument(key, value)
	t.NoError(err)

	b, err := Serialize(origDoc)
	t.NoError(err)
	t.NotEmpty(b)

	var doc *Document
	err = Deserialize(b, &doc)
	t.NoError(err)

	t.Equal(origDoc.K, doc.K)
	t.NotEqual(origDoc.V, doc.V)

	var valueMap map[string]uint64
	err = doc.Decode(&valueMap)
	t.NoError(err)
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
		b, err := Serialize(doc)
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
		t.Equal(value, returned)

		for k, v := range returned {
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
		b, err := Serialize(doc)
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
		check func(interface{})
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
			input: Document{},
			err:   nil,
			t:     Struct,
			m:     bsontype.EmbeddedDocument,
			msg:   "Document{}",
			check: func(e interface{}) {
				d := Document{}
				t.True(d.Equal(e.(Document)))
			},
		},
		{
			input: &Document{},
			err:   nil,
			t:     Ptr,
			m:     bsontype.EmbeddedDocument,
			msg:   "&Document{}",
			check: func(e interface{}) {
				d := &Document{}
				u := e.(*Document)
				t.True(d.Equal(*u))
			},
		},
	}

	for n, c := range cases {
		bt, encoded, err := convertToDocumentValue(c.input)
		if c.err == nil {
			t.NoError(err, c.msg)
		} else {
			t.EqualError(c.err, err.Error(), c.msg)
		}

		t.Equal(c.m, bt, c.msg)

		if c.err != nil || c.m == bsontype.Undefined {
			continue
		}

		b, err := Serialize(bson.M{"a": "showme", "b": encoded})
		t.NoError(err, "%s: %d", c.msg, n)

		raw := bson.Raw(b)
		rv := raw.Lookup("b")

		f := reflect.New(reflect.TypeOf(c.input)).Interface()
		err = DecodeDocumentValue(rv, f)
		t.NoError(err, "%s: %d", c.msg, n)

		e := reflect.ValueOf(f).Elem().Interface()
		t.Equal(c.t, Hint(reflect.TypeOf(e).Kind()), "%s: %d", c.msg, n)

		if c.check != nil {
			c.check(e)
		}
	}
}

func (t *testMongoValue) TestMarshalStruct() {
	input := struct {
		A string `bson:"a"`
		B []int  `bson:"b"`
	}{A: "showme", B: []int{1, 2}}

	b, err := Serialize(input)
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

type testMongoDocumentStore struct {
	baseTestMongoStorage
}

func (t *testMongoDocumentStore) TestFindMap() {
	key := "showme"
	value := map[string]uint64{
		"a1": 1,
		"a2": 2,
	}

	err := t.s.Insert(key, value)
	t.NoError(err)

	{
		var record map[string]uint64
		err := t.s.Get(key, &record)
		t.NoError(err)
	}

	cur, err := t.s.Collection().Find(context.Background(), bson.M{DOC.Field("a1"): 1})
	t.NoError(err)

	var records []map[string]uint64
	for cur.Next(context.Background()) {
		var value map[string]uint64
		_, err := UnmarshalDocument([]byte(cur.Current), &value)
		t.NoError(err)

		records = append(records, value)
	}
	err = cur.Err()
	t.NoError(err)

	t.Equal(1, len(records))

	for k, v := range records[0] {
		t.Equal(value[k], v)
	}
}

func (t *testMongoDocumentStore) TestFindStruct() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r)
		}
	}()

	key := "showme"
	value := testUnmarshalStruct{
		A: "AAA",
		B: 99,
		C: []uint64{7, 8, 9},
	}

	err := t.s.Insert(key, value)
	t.NoError(err)

	for i := 0; i < 5; i++ {
		key := common.RandomUUID()
		value := testUnmarshalStruct{
			A: common.RandomUUID(),
			B: 199,
			C: []uint64{7, 8, 9},
		}
		err := t.s.Insert(key, value)
		t.NoError(err)
	}

	{
		var record testUnmarshalStruct
		err := t.s.Get(key, &record)
		t.NoError(err)
	}

	compare := func(records []testUnmarshalStruct) {
		t.Equal(1, len(records))

		t.Equal(value.A, records[0].A)
		t.Equal(value.B, records[0].B)
		t.Equal(value.C, records[0].C)
	}

	{
		cur, err := t.s.Collection().Find(context.Background(), bson.M{"_v.a": "AAA"})
		t.NoError(err)

		var records []testUnmarshalStruct
		for cur.Next(context.Background()) {
			var value testUnmarshalStruct
			_, err := UnmarshalDocument([]byte(cur.Current), &value)
			t.NoError(err)

			records = append(records, value)
		}
		err = cur.Err()
		t.NoError(err)

		compare(records)
	}

	{
		cur, err := t.s.Collection().Find(context.Background(), bson.M{"_v.b": 99})
		t.NoError(err)

		var records []testUnmarshalStruct
		for cur.Next(context.Background()) {
			var value testUnmarshalStruct
			_, err := UnmarshalDocument([]byte(cur.Current), &value)
			t.NoError(err)

			records = append(records, value)
		}
		err = cur.Err()
		t.NoError(err)

		compare(records)
	}

	{
		cur, err := t.s.Collection().Find(
			context.Background(),
			bson.M{"_v.a": "AAA", "_v.c": bson.M{"$in": bson.A{7}}},
		)
		t.NoError(err)

		var records []testUnmarshalStruct
		for cur.Next(context.Background()) {
			var value testUnmarshalStruct
			_, err := UnmarshalDocument([]byte(cur.Current), &value)
			t.NoError(err)

			records = append(records, value)
		}
		err = cur.Err()
		t.NoError(err)

		compare(records)
	}
}

func TestMongoValue(t *testing.T) {
	suite.Run(t, new(testMongoValue))
}

func TestMongoDocumentStore(t *testing.T) {
	if client, err := connect(); err != nil {
		log.Warn("mongodb test will be skipped")
		return
	} else {
		disconnect(client)
	}

	suite.Run(t, new(testMongoDocumentStore))
}
