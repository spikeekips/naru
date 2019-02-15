package storage

import (
	observable "github.com/GianlucaGuarini/go-observable"
)

const (
	EventNewBlock      string = "\x00\x01"
	EventNewAccount    string = "\x00\x02"
	EventUpdateAccount string = "\x00\x03"
	EventNewItem       string = "\x00\x04"
)

var (
	Observer *Observable
)

func init() {
	Observer = NewObservable(observable.New())
}

type EventDBCore interface {
	Events() map[string][]interface{}
	AddEvent(string, ...interface{})
	ClearEvents() map[string][]interface{}
}

type Observable struct {
	*observable.Observable
}

func NewObservable(ob *observable.Observable) *Observable {
	return &Observable{Observable: ob}
}

func (o *Observable) On(event string, callback interface{}) *Observable {
	o.Observable.On(event, callback)
	return o
}

func (o *Observable) One(event string, callback interface{}) *Observable {
	o.Observable.One(event, callback)
	return o
}

func (o *Observable) Off(event string, args ...interface{}) *Observable {
	o.Observable.Off(event, args...)
	return o
}

func (o *Observable) Trigger(event string, v ...interface{}) *Observable {
	o.Observable.Trigger(event, v...)

	return o
}
