package leveldbstorage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/storage"
)

type testLevelDBStorage struct {
	suite.Suite
	s *Storage
}

func (t *testLevelDBStorage) SetupTest() {
	var c *config.LevelDBStorage
	{
		c = &config.LevelDBStorage{Path: "memory://"}
		err := c.Validate()
		t.NoError(err)
	}

	s, err := NewStorage(c)
	t.NoError(err)

	t.s = s
}

func (t *testLevelDBStorage) TeardownTest() {
	t.s.Close()
}

func (t *testLevelDBStorage) TestInsertAndGet() {
	key := "showme"
	value := "findme"
	{
		err := t.s.Insert(key, value)
		t.NoError(err)
	}

	var returned string
	err := t.s.Get(key, &returned)
	t.NoError(err)
	t.Equal(value, returned)
}

func (t *testLevelDBStorage) TestGetUnknownKey() {
	var returned string
	err := t.s.Get("whoareyou?", &returned)
	t.True(storage.NotFound.Equal(err))
}

func (t *testLevelDBStorage) TestIterator() {
	var inserted []storage.Record
	for i := 0; i < 20; i++ {
		item := storage.Record{
			Key:   fmt.Sprintf("item-%02d", i),
			Value: []byte(common.RandomUUID()),
		}
		err := t.s.Insert(item.Key, item.Value)
		t.NoError(err)

		inserted = append(inserted, item)
	}

	var limit uint64
	{
		limit = 5
		options := storage.NewDefaultListOptions(false, nil, limit)
		iter, cls := t.s.Iterator("item", []byte{}, options)
		defer cls()

		var items []storage.Record
		for {
			item, next := iter()
			if !next {
				break
			}
			items = append(items, item)
		}

		t.Equal(int(limit), len(items))

		for i, item := range items {
			t.Equal(inserted[i].Key, item.Key)
			t.Equal(inserted[i].Value, item.Value)
		}
	}

	{
		limit = uint64(len(inserted) + 10)
		options := storage.NewDefaultListOptions(false, nil, limit)
		iter, cls := t.s.Iterator("item", []byte{}, options)
		defer cls()

		var items []storage.Record
		for {
			item, next := iter()
			if len(items) == len(inserted) {
				t.False(next)
			}
			if !next {
				break
			}
			items = append(items, item)
		}

		t.Equal(len(inserted), len(items))

		for i, item := range items {
			t.Equal(inserted[i].Key, item.Key)
			t.Equal(inserted[i].Value, item.Value)
		}
	}
}

func (t *testLevelDBStorage) TestIteratorEmpty() {
	var limit uint64
	{
		limit = 5
		options := storage.NewDefaultListOptions(false, nil, limit)
		iter, cls := t.s.Iterator("item", []byte{}, options)
		defer cls()

		var items []storage.Record
		for {
			item, next := iter()
			if !next {
				break
			}
			items = append(items, item)
		}

		t.Equal(0, len(items))
	}
}

func (t *testLevelDBStorage) TestBatch() {
	batch := t.s.Batch()

	key := "showme"
	value := "findme"
	{
		err := batch.Insert(key, value)
		t.NoError(err)
	}

	{
		err := t.s.Get(key, nil)
		t.True(storage.NotFound.Equal(err))
	}

	err := batch.Write()
	t.NoError(err)

	{
		var returned string
		err := t.s.Get(key, &returned)
		t.NoError(err)
		t.Equal(value, returned)
	}
}

func (t *testLevelDBStorage) TestEvent() {
	event := common.RandomUUID()
	defer storage.Observer.Off(event)

	evented := make(chan interface{})
	done := make(chan bool)
	defer close(evented)
	defer close(done)

	storage.Observer.On(event, func(args ...interface{}) {
		evented <- args
	})

	key := "showme"
	value := "findme"

	go func() {
		for {
			select {
			case fired := <-evented:
				if fired == nil {
					return
				}
				t.Equal(value, fired.([]interface{})[0].(string))
				done <- true
			}
		}
	}()

	err := t.s.Insert(key, value)
	t.NoError(err)
	t.s.Event(event, value)

	select {
	case <-done:
		break
	}
}

func (t *testLevelDBStorage) TestBatchEventWillNotTriggerBeforeWrite() {
	event := common.RandomUUID()
	defer storage.Observer.Off(event)

	evented := make(chan interface{}, 1)
	fired := make(chan interface{}, 1)
	defer close(evented)
	defer close(fired)

	storage.Observer.On(event, func(args ...interface{}) {
		evented <- args
	})

	key := "showme"
	value := "findme"

	go func() {
		select {
		case a, ok := <-evented:
			if !ok {
				return
			}
			fired <- a
		}
	}()

	batch := t.s.Batch()
	{
		err := batch.Insert(key, value)
		t.NoError(err)
		batch.Event(event, value)
	}

	select {
	case <-time.After(time.Second * 1):
		break
	case a := <-fired:
		t.Empty(a)
	}
}

func (t *testLevelDBStorage) TestBatchEventTriggerAfterWrite() {
	event := common.RandomUUID()
	defer storage.Observer.Off(event)

	evented := make(chan interface{})
	done := make(chan bool)
	defer close(evented)
	defer close(done)

	storage.Observer.On(event, func(args ...interface{}) {
		evented <- args
	})

	key := "showme"
	value := "findme"

	go func() {
		select {
		case fired := <-evented:
			t.Equal(value, fired.([]interface{})[0].(string))
			done <- true
		}
	}()

	batch := t.s.Batch()
	err := batch.Insert(key, value)
	t.NoError(err)
	batch.Event(event, value)

	err = batch.Write()
	t.NoError(err)

	select {
	case <-done:
		break
	}
}

func (t *testLevelDBStorage) TestBatchEventTriggerSynchronous() {
	event := common.RandomUUID()
	defer storage.Observer.Off(event)

	evented := make(chan int)
	done := make(chan []int)
	defer close(evented)
	defer close(done)

	storage.Observer.On(event, func(args ...interface{}) {
		time.Sleep(time.Second * 1)
		evented <- 0
	})

	go func() {
		var seq []int
	end:
		for {
			select {
			case e := <-evented:
				seq = append(seq, e)
				if len(seq) == 2 {
					done <- seq
					break end
				}
			}
		}
	}()

	batch := t.s.Batch()
	err := batch.Insert("showme", "findme")
	t.NoError(err)
	batch.Event(event, "findme")

	err = batch.Write()
	evented <- 1
	t.NoError(err)

	select {
	case seq := <-done:
		t.Equal(0, seq[0])
		t.Equal(1, seq[1])
		break
	}
}

func (t *testLevelDBStorage) TestBatchEventTriggerAsynchronous() {
	event := common.RandomUUID()
	defer storage.Observer.Off(event)

	evented := make(chan int)
	done := make(chan []int)
	defer close(evented)
	defer close(done)

	storage.Observer.Async(event, func(args ...interface{}) {
		time.Sleep(time.Second * 1)
		evented <- 0
	})

	go func() {
		var seq []int
	end:
		for {
			select {
			case e := <-evented:
				seq = append(seq, e)
				if len(seq) == 2 {
					done <- seq
					break end
				}
			}
		}
	}()

	batch := t.s.Batch()
	err := batch.Insert("showme", "findme")
	t.NoError(err)
	batch.Event(event, "findme")

	err = batch.Write()
	evented <- 1
	t.NoError(err)

	select {
	case seq := <-done:
		t.Equal(1, seq[0])
		t.Equal(0, seq[1])
		break
	}
}

func TestLevelDBStorage(t *testing.T) {
	suite.Run(t, new(testLevelDBStorage))
}
