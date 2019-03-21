package common

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testObservable struct {
	suite.Suite
}

func (t *testObservable) TestSync() {
	o := NewObservable(RandomUUID())
	event := newRandomEvent()

	var evented string
	o.Sync(event, func(a string) {
		evented = a
	})

	arg := "findme"
	o.Trigger(event, arg)
	t.Equal(arg, evented)
}

func (t *testObservable) TestAsync() {
	o := NewObservable(RandomUUID())
	event := newRandomEvent()

	evented := make(chan string)
	defer close(evented)

	o.Async(event, func(a string) {
		evented <- a
	})

	arg := "findme"
	o.Trigger(event, arg)

	select {
	case e := <-evented:
		t.Equal(arg, e)
	}
}

func (t *testObservable) TestAsyncOff() {
	o := NewObservable(RandomUUID())
	event := newRandomEvent()

	evented := make(chan string)
	defer close(evented)

	o.Async(event, func(a string) {
		evented <- a
	})

	arg := "findme"
	o.Trigger(event, arg)

	select {
	case e := <-evented:
		t.Equal(arg, e)
	}

	t.True(func() bool {
		_, found := o.async[event]
		return found
	}())

	// after `Off()`, async event will be removed
	o.Off(event, nil)
	t.False(func() bool {
		_, found := o.async[event]
		return found
	}())
}

func TestObservable(t *testing.T) {
	suite.Run(t, new(testObservable))
}

func newRandomEvent() string {
	return RandomUUID()
}
