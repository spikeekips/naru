package mongostorage

import (
	"encoding/base64"
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
	tUint64      reflect.Type = reflect.TypeOf(uint64(0))
	tSEBAKAmount reflect.Type = reflect.TypeOf(sebakcommon.Amount(0))
	tBytes       reflect.Type = reflect.TypeOf([]byte{})
)

var DefaultBSONRegistry *bsoncodec.Registry = NewBSONRegistry()

func Uint64EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUint64 {
		return bsoncodec.ValueEncoderError{Name: "Uint64EncodeValue", Types: []reflect.Type{tUint64}, Received: val}
	}

	n, err := primitive.ParseDecimal128(strconv.FormatUint(val.Interface().(uint64), 10))
	if err != nil {
		return err
	}
	return vw.WriteDecimal128(n)
}

func Uint64DecodeValue(dctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if vr.Type() != bsontype.Decimal128 {
		return fmt.Errorf("cannot decode %v into sebakcommon.Amount", vr.Type())
	}

	if !val.CanSet() || val.Type() != tUint64 {
		return bsoncodec.ValueDecoderError{Name: "Uint64DecodeValue", Types: []reflect.Type{tUint64}, Received: val}
	}

	d128, err := vr.ReadDecimal128()
	if err != nil {
		return err
	}
	c, err := strconv.ParseUint(d128.String(), 10, 64)
	if err != nil {
		return err
	}

	val.Set(reflect.ValueOf(c))
	return nil
}

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

func BytesEncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tBytes {
		return bsoncodec.ValueEncoderError{Name: "BytesEncodeValue", Types: []reflect.Type{tBytes}, Received: val}
	}

	return vw.WriteString(
		base64.StdEncoding.EncodeToString(val.Interface().([]byte)),
	)
}

func BytesDecodeValue(dctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if vr.Type() != bsontype.String {
		return fmt.Errorf("cannot decode %v into []byte", vr.Type())
	}

	if !val.CanSet() || val.Type() != tBytes {
		return bsoncodec.ValueDecoderError{Name: "BytesDecodeValue", Types: []reflect.Type{tBytes}, Received: val}
	}

	s, err := vr.ReadString()
	if err != nil {
		return err
	}
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	val.Set(reflect.ValueOf(data))
	return nil
}

func NewBSONRegistry() *bsoncodec.Registry {
	rb := bson.NewRegistryBuilder()
	rb.RegisterEncoder(tSEBAKAmount, bsoncodec.ValueEncoderFunc(SEBAKAmountEncodeValue))
	rb.RegisterDecoder(tSEBAKAmount, bsoncodec.ValueDecoderFunc(SEBAKAmountDecodeValue))
	rb.RegisterEncoder(tUint64, bsoncodec.ValueEncoderFunc(Uint64EncodeValue))
	rb.RegisterDecoder(tUint64, bsoncodec.ValueDecoderFunc(Uint64DecodeValue))
	rb.RegisterEncoder(tBytes, bsoncodec.ValueEncoderFunc(BytesEncodeValue))
	rb.RegisterDecoder(tBytes, bsoncodec.ValueDecoderFunc(BytesDecodeValue))

	return rb.Build()
}
