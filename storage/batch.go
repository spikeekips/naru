package storage

import (
	"sync"

	sebakstorage "boscoin.io/sebak/lib/storage"
)

type BatchCore struct {
	sync.RWMutex
	*sebakstorage.BatchCore
	events map[string][]interface{}
}

func NewBatchCore(bc *sebakstorage.BatchCore) *BatchCore {
	return &BatchCore{
		BatchCore: bc,
		events:    map[string][]interface{}{},
	}
}

func (bc *BatchCore) Events() map[string][]interface{} {
	bc.RLock()
	defer bc.RUnlock()

	return bc.events
}

func (bc *BatchCore) AddEvent(event string, v ...interface{}) {
	bc.Lock()
	defer bc.Unlock()

	bc.events[event] = append(bc.events[event], v...)
}

func (bc *BatchCore) ClearEvents() map[string][]interface{} {
	bc.Lock()
	defer bc.Unlock()

	events := map[string][]interface{}{}
	for k, v := range bc.events {
		events[k] = v
	}

	bc.events = map[string][]interface{}{}

	return events
}
