package common

import (
	"encoding/json"
)

const (
	CommonErrorCode = iota + 100
	HTTPProblemCode
)

var (
	CommonError = NewError(CommonErrorCode, "")
	HTTPProblem = NewError(HTTPProblemCode, "http problem")
)

type Error struct {
	code    uint                   `json:"code"`
	message string                 `json:"message"`
	data    map[string]interface{} `json:"data"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e *Error) Code() uint {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) SetMessage(m string) *Error {
	e.message = m

	return e
}

func (e *Error) Data() map[string]interface{} {
	return e.data
}

func (e *Error) Equal(n error) bool {
	ne, found := n.(*Error)
	if found {
		return e.Code() == ne.Code()
	}

	return false
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"code":    e.code,
		"message": e.message,
		"data":    e.data,
	})
}

func (e *Error) SetData(k string, v interface{}) *Error {
	e.data[k] = v

	return e
}

func (e *Error) New() *Error {
	var new Error
	new = *e

	new.data = map[string]interface{}{}
	if e.data != nil && len(e.data) > 0 {
		for k, v := range e.data {
			new.data[k] = v
		}
	}

	return &new
}

func (e *Error) NewFromError(err error) *Error {
	return NewError(e.Code(), err.Error())
}

func NewError(code uint, message string) *Error {
	return &Error{code: code, message: message, data: map[string]interface{}{}}
}
