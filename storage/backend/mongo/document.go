package mongostorage

import (
	"reflect"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Document struct {
	K  string      `bson:"_k"`
	V  interface{} `bson:"_v"`
	rv bson.RawValue
}

func NewDocument(key string, value interface{}) (Document, error) {
	raw, err := Serialize(bson.M{"_v": value})
	if err != nil {
		return Document{}, err
	}

	return Document{K: key, V: value, rv: bson.Raw(raw).Lookup("_v")}, nil
}

func (d Document) Key() string {
	return d.K
}

func (d Document) Value() interface{} {
	return d.V
}

func (d Document) Decode(v interface{}) error {
	return d.rv.UnmarshalWithRegistry(DefaultBSONRegistry, v)
}

func (d Document) MarshalBSON() ([]byte, error) {
	return Serialize(bson.M{"_k": d.K, "_v": d.V})
}

func (d *Document) UnmarshalBSON(b []byte) error {
	var m bson.M
	if err := Deserialize(b, &m); err != nil {
		return err
	}

	*d = Document{
		K:  m["_k"].(string),
		V:  m["_v"],
		rv: bson.Raw(b).Lookup("_v"),
	}

	return nil
}

func (d Document) Equal(n Document) bool {
	return d.K == n.K && reflect.DeepEqual(d.V, n.V)
}

func decimal128ToUint(d primitive.Decimal128, bitSize int) (interface{}, error) {
	return strconv.ParseUint(d.String(), 10, bitSize)
}

func UnmarshalDocument(b []byte, v interface{}) (Document, error) {
	var doc Document
	if err := Deserialize(b, &doc); err != nil {
		return Document{}, err
	}

	return doc, doc.Decode(v)
}
