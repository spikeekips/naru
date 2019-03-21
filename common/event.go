package common

import (
	"strings"

	observable "github.com/GianlucaGuarini/go-observable"
	logging "github.com/inconshreveable/log15"
)

type Observable struct {
	name string
	*observable.Observable
	async map[string]struct{}
	log   logging.Logger
}

func NewObservable(name string) *Observable {
	o := &Observable{
		Observable: observable.New(),
		name:       name,
		log:        log.New(logging.Ctx{"module": "event", "name": name}),
		async:      map[string]struct{}{},
	}

	return o
}

func (o *Observable) Name() string {
	return o.name
}

func (o *Observable) on(events string, callback interface{}) *Observable {
	o.Observable.On(events, callback)
	return o
}

func (o *Observable) Sync(events string, callback interface{}) *Observable {
	return o.on(events, callback)
}

func (o *Observable) Async(events string, callback interface{}) *Observable {
	o.Lock()
	for _, e := range strings.Fields(events) {
		o.async[e] = struct{}{}
	}
	o.Unlock()

	return o.on(events, callback)
}

func (o *Observable) One(events string, callback interface{}) *Observable {
	o.Observable.One(events, callback)
	return o
}

func (o *Observable) Off(events string, args ...interface{}) *Observable {
	o.Lock()
	for _, e := range strings.Fields(events) {
		if o.isAsync(e) {
			delete(o.async, e)
		}
	}
	o.Unlock()

	o.Observable.Off(events, args...)
	return o
}

func (o *Observable) Trigger(events string, args ...interface{}) *Observable {
	var async, sync []string
	for _, e := range strings.Fields(events) {
		if o.isAsync(e) {
			async = append(async, e)
		} else {
			sync = append(sync, e)
		}
	}

	if len(async) > 0 {
		go func() {
			o.Observable.Trigger(strings.Join(async, " "), args...)
		}()
	}

	if len(sync) > 0 {
		o.Observable.Trigger(strings.Join(sync, " "), args...)
	}

	return o
}

func (o *Observable) isAsync(event string) bool {
	_, found := o.async[event]
	return found
}

type EventItem struct {
	Events string
	Items  []interface{}
}

func NewEventItem(events string, items ...interface{}) EventItem {
	return EventItem{Events: events, Items: items}
}
