package common

import (
	"bytes"
	"encoding/json"
)

func marshalJSONNotEscapeHTML(o interface{}, indent bool) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(indent)
	err := encoder.Encode(o)
	return bytes.TrimSpace(buffer.Bytes()), err
}

func MarshalJSONNotEscapeHTMLIndent(o interface{}) ([]byte, error) {
	return marshalJSONNotEscapeHTML(o, true)
}

func MarshalJSONNotEscapeHTML(o interface{}) ([]byte, error) {
	return marshalJSONNotEscapeHTML(o, false)
}
