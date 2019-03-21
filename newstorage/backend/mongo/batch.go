package mongostorage

import (
	"context"
	"sync"

	logging "github.com/inconshreveable/log15"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/newstorage"
)

type Batch struct {
	sync.RWMutex
	s      *Storage
	ops    []mongo.WriteModel
	events []common.EventItem
	log    logging.Logger
}

func NewBatch(s *Storage) *Batch {
	return &Batch{
		s:   s,
		log: log.New(logging.Ctx{"name": "batch"}),
	}
}

func (b *Batch) Core() *mongo.Client {
	return b.s.Core()
}

func (b *Batch) Close() error {
	return b.Cancel()
}

func (b *Batch) Batch() newstorage.BatchStorage {
	return b
}

func (b *Batch) Write() error {
	b.RLock()
	result, err := b.s.Collection().BulkWrite(context.Background(), b.ops)
	if err != nil {
		b.RUnlock()
		return err
	}
	b.RUnlock()

	b.log.Debug(
		"write",
		"inserted", result.InsertedCount,
		"matched", result.MatchedCount,
		"modified", result.ModifiedCount,
		"deleted", result.DeletedCount,
		"upserted", result.UpsertedCount,
		"object-ids", result.UpsertedIDs,
	)

	b.RLock()
	for _, e := range b.events {
		newstorage.Observer.Trigger(e.Events, e.Items...)
	}
	b.RUnlock()

	b.clearEvents()
	return nil
}

func (b *Batch) Cancel() error {
	defer b.clearEvents()
	b.ops = nil

	return nil
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

func (b *Batch) Iterator(prefix string, v interface{}, options newstorage.ListOptions) (func() (newstorage.Record, bool), func()) {
	return b.s.Iterator(prefix, v, options)
}

func (b *Batch) Insert(k string, v interface{}) error {
	if err := b.MustNotExist(k); err != nil {
		return err
	}

	return b.insert(k, v)
}

func (b *Batch) insert(k string, v interface{}) error {
	b.Lock()
	defer b.Unlock()

	doc, err := NewDocument(resolveKey(k), v)
	if err != nil {
		return err
	}

	b.ops = append(b.ops, mongo.NewInsertOneModel().SetDocument(doc))

	return nil
}

func (b *Batch) Update(k string, v interface{}) error {
	if err := b.MustExist(k); err != nil {
		return err
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
	if err := b.MustExist(k); err != nil {
		return err
	}

	return b.delete(k)
}

func (b *Batch) delete(k string) error {
	b.Lock()
	defer b.Unlock()

	b.ops = append(b.ops, mongo.NewDeleteOneModel().SetFilter(bson.M{"_k": resolveKey(k)}))
	log.Debug("delete doc", "key", k)

	return nil
}

func (b *Batch) MultipleInsert(items ...newstorage.Value) error {
	return nil
}

func (b *Batch) MultipleUpdate(items ...newstorage.Value) error {
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
