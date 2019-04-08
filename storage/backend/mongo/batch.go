package mongostorage

import (
	"context"
	"strings"
	"sync"

	logging "github.com/inconshreveable/log15"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/storage"
)

type Batch struct {
	sync.RWMutex
	s          *Storage
	ops        map[string][]mongo.WriteModel
	events     []common.EventItem
	log        logging.Logger
	skipExists bool
}

func NewBatch(s *Storage, skipExists bool) (*Batch, error) {
	ns, err := s.New()
	if err != nil {
		return nil, err
	}

	return &Batch{
		s:          ns,
		log:        log.New(logging.Ctx{"name": "batch"}),
		ops:        map[string][]mongo.WriteModel{},
		skipExists: skipExists,
	}, nil
}

func (b *Batch) Core() *mongo.Client {
	return b.s.Core()
}

func (b *Batch) Close() error {
	b.clearEvents()

	b.Lock()
	b.ops = map[string][]mongo.WriteModel{}
	b.Unlock()

	return b.s.Close()
}

func (b *Batch) Initialize() error {
	return nil
}

func (b *Batch) Batch() (storage.BatchStorage, error) {
	return b, nil
}

func (b *Batch) Write() (err error) {
	defer func() {
		if err == nil {
			return
		}

		b.log.Error("failed to BulkWrite", "error", err)
	}()

	defer b.Close()

	var events []common.EventItem
	{
		b.RLock()
		events = make([]common.EventItem, len(b.events))
		copy(events, b.events)
		b.RUnlock()
	}

	for _, e := range events { // only OnAfterSave event will be triggered
		var es []string
		for _, n := range strings.Fields(e.Events) {
			if !strings.HasPrefix(n, "OnAfterSave") {
				continue
			}
			es = append(es, n)
		}

		storage.Observer.Trigger(strings.Join(es, " "), e.Items...)
	}

	b.RLock()
	ops := map[string][]mongo.WriteModel{}
	for k, v := range b.ops {
		ops[k] = make([]mongo.WriteModel, len(v))
		copy(ops[k], v)
	}
	b.RUnlock()

	var result *mongo.BulkWriteResult
	for k, v := range ops {
		col := b.s.Database().Collection(k)
		result, err = col.BulkWrite(context.Background(), v)
		if err != nil {
			return
		}

		b.log.Debug(
			"write",
			"collection", k,
			"inserted", result.InsertedCount,
			"matched", result.MatchedCount,
			"modified", result.ModifiedCount,
			"deleted", result.DeletedCount,
			"upserted", result.UpsertedCount,
			"object-ids", result.UpsertedIDs,
		)
	}

	for _, e := range events { // only non-OnAfterSave event will be triggered
		var events []string
		for _, n := range strings.Fields(e.Events) {
			if strings.HasPrefix(n, "OnAfterSave") {
				continue
			}
			events = append(events, n)
		}

		storage.Observer.Trigger(strings.Join(events, " "), e.Items...)
	}

	return
}

func (b *Batch) Cancel() error {
	return b.Close()
}

func (b *Batch) Has(k string) (bool, error) {
	return b.s.Has(k)
}

func (b *Batch) MustExist(k string) error {
	return b.s.MustExist(k)
}

func (b *Batch) MustNotExist(k string) error {
	return b.s.MustNotExist(k)
}

func (b *Batch) Get(k string, v interface{}) error {
	return b.s.Get(k, v)
}

func (b *Batch) Iterator(prefix string, v interface{}, options storage.ListOptions) (func() (storage.Record, bool), func()) {
	return b.s.Iterator(prefix, v, options)
}

func (b *Batch) Insert(k string, v interface{}) error {
	if !b.skipExists {
		if err := b.MustNotExist(k); err != nil {
			return err
		}
	}

	return b.insert(k, v)
}

func (b *Batch) insert(k string, v interface{}) error {
	c, err := getCollection(k)
	if err != nil {
		return err
	}

	b.Lock()
	if _, ok := b.ops[c]; !ok {
		b.ops[c] = []mongo.WriteModel{}
	}
	b.Unlock()

	doc, err := NewDocument(k, v)
	if err != nil {
		return err
	}

	b.Lock()
	b.ops[c] = append(b.ops[c], mongo.NewInsertOneModel().SetDocument(doc))
	b.Unlock()

	return nil
}

func (b *Batch) Update(k string, v interface{}) error {
	if !b.skipExists {
		if err := b.MustExist(k); err != nil {
			return err
		}
	}

	if err := b.delete(k); err != nil {
		return err
	}
	if err := b.insert(k, v); err != nil {
		return err
	}
	return nil
}

func (b *Batch) Delete(k string) error {
	if !b.skipExists {
		if err := b.MustExist(k); err != nil {
			return err
		}
	}

	return b.delete(k)
}

func (b *Batch) delete(k string) error {
	c, err := getCollection(k)
	if err != nil {
		return err
	}

	b.Lock()
	defer b.Unlock()
	if _, ok := b.ops[c]; !ok {
		b.ops[c] = []mongo.WriteModel{}
	}

	b.ops[c] = append(b.ops[c], mongo.NewDeleteOneModel().SetFilter(bson.M{KEY: k}))
	log.Debug("delete doc", "key", k)

	return nil
}

func (b *Batch) MultipleInsert(items ...storage.Value) error {
	return nil
}

func (b *Batch) MultipleUpdate(items ...storage.Value) error {
	return nil
}

func (b *Batch) MultipleDelete(keys ...string) error {
	return nil
}

func (b *Batch) Event(event string, values ...interface{}) {
	b.Lock()
	defer b.Unlock()

	b.events = append(b.events, common.NewEventItem(event, values...))
	return
}

func (b *Batch) clearEvents() {
	b.Lock()
	defer b.Unlock()

	b.events = nil
}
