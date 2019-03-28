package element

import (
	"errors"
	"reflect"

	"github.com/spikeekips/naru/common"
)

func SyncWithError(fns ...interface{}) func(*common.Error, ...interface{}) {
	for _, fn := range fns {
		t := reflect.TypeOf(fn)
		if t.Kind() != reflect.Func {
			panic(errors.New("invalid event function found: not function"))
		}

		if t.NumOut() != 1 {
			panic(errors.New("invalid event function found: only one value should be returned"))
		}

		if t.Out(0).Name() != "error" {
			panic(errors.New("invalid event function found: return type is not error"))
		}
	}

	return func(err *common.Error, args ...interface{}) {
		var rValues []reflect.Value
		for _, arg := range args {
			rValues = append(rValues, reflect.ValueOf(arg))
		}

		for _, fn := range fns {
			rs := reflect.ValueOf(fn).Call(rValues)

			e := rs[0].Interface()
			if e == nil {
				continue
			}

			ce, found := e.(*common.Error)
			if !found {
				ce = common.CommonError.New().SetMessage(ce.Error())
			}
			*err = *ce
			return
		}
	}
}
