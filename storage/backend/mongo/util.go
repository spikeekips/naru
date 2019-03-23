package mongostorage

import (
	"fmt"
	"reflect"
	"strconv"

	sebakcommon "boscoin.io/sebak/lib/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	tSEBAKAmount reflect.Type = reflect.TypeOf(sebakcommon.Amount(0))
)

var DefaultBSONRegistry *bsoncodec.Registry = NewBSONRegistry()

func SEBAKAmountEncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tSEBAKAmount {
		return bsoncodec.ValueEncoderError{Name: "SEBAKAmountEncodeValue", Types: []reflect.Type{tSEBAKAmount}, Received: val}
	}

	n, err := primitive.ParseDecimal128(val.Interface().(sebakcommon.Amount).String())
	if err != nil {
		return err
	}
	return vw.WriteDecimal128(n)
}

func SEBAKAmountDecodeValue(dctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if vr.Type() != bsontype.Decimal128 {
		return fmt.Errorf("cannot decode %v into sebakcommon.Amount", vr.Type())
	}

	if !val.CanSet() || val.Type() != tSEBAKAmount {
		return bsoncodec.ValueDecoderError{Name: "SEBAKAmountDecodeValue", Types: []reflect.Type{tSEBAKAmount}, Received: val}
	}

	d128, err := vr.ReadDecimal128()
	if err != nil {
		return err
	}
	c, err := strconv.ParseUint(d128.String(), 10, 64)
	if err != nil {
		return err
	}

	val.Set(reflect.ValueOf(sebakcommon.Amount(c)))
	return nil
}

func NewBSONRegistry() *bsoncodec.Registry {
	rb := bson.NewRegistryBuilder()
	rb.RegisterEncoder(tSEBAKAmount, bsoncodec.ValueEncoderFunc(SEBAKAmountEncodeValue))
	rb.RegisterDecoder(tSEBAKAmount, bsoncodec.ValueDecoderFunc(SEBAKAmountDecodeValue))

	return rb.Build()
}

func Serialize(i interface{}) ([]byte, error) {
	return bson.MarshalWithRegistry(DefaultBSONRegistry, i)
}

func Deserialize(b []byte, i interface{}) error {
	return bson.UnmarshalWithRegistry(DefaultBSONRegistry, b, i)
}
