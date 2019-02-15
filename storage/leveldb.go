package storage

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDBCore struct {
	sync.RWMutex
	*leveldb.DB
	events map[string][]interface{}
}

func NewLevelDBCore(db *leveldb.DB) *LevelDBCore {
	return &LevelDBCore{
		DB:     db,
		events: map[string][]interface{}{},
	}
}

func (lc *LevelDBCore) Events() map[string][]interface{} {
	lc.RLock()
	defer lc.RUnlock()

	return lc.events
}

func (lc *LevelDBCore) AddEvent(event string, v ...interface{}) {
	lc.Lock()
	defer lc.Unlock()

	lc.events[event] = append(lc.events[event], v...)
}

func (lc *LevelDBCore) ClearEvents() map[string][]interface{} {
	lc.Lock()
	defer lc.Unlock()

	events := map[string][]interface{}{}
	for k, v := range lc.events {
		events[k] = v
	}

	lc.events = map[string][]interface{}{}

	return events
}
