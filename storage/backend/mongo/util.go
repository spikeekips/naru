package mongostorage

import (
	"go.mongodb.org/mongo-driver/bson"
)

func Serialize(i interface{}) ([]byte, error) {
	return bson.MarshalWithRegistry(DefaultBSONRegistry, i)
}

func Deserialize(b []byte, i interface{}) error {
	return bson.UnmarshalWithRegistry(DefaultBSONRegistry, b, i)
}
