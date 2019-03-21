package newstorage

import (
	"encoding"
	"encoding/json"
)

// Encapsulate serialization method for various functions
func Serialize(i interface{}) ([]byte, error) {
	if bm, ok := i.(encoding.BinaryMarshaler); ok {
		return bm.MarshalBinary()
	}
	return json.Marshal(&i)
}

// Encapsulate deserialization method for various functions
func Deserialize(data []byte, i interface{}) error {
	if bm, ok := i.(encoding.BinaryUnmarshaler); ok {
		return bm.UnmarshalBinary(data)
	}
	return json.Unmarshal(data, &i)
}
