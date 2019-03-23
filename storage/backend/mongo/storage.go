package mongostorage

import (
	"context"
	"encoding/hex"
	"reflect"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/storage"
)

var defaultCollectionName string = "default"

type Storage struct {
	c              *mongo.Client
	databaseName   string
	collectionName string
}

func NewStorage(c *config.MongoStorage) (*Storage, error) {
	client, err := mongo.NewClient(c.URI)
	if err != nil {
		return nil, err
	}

	s := &Storage{
		c:              client,
		databaseName:   c.DB,
		collectionName: defaultCollectionName,
	}
	if err := s.connect(); err != nil {
		return nil, err
	}

	return s, nil
}

func (b *Storage) connect() error {
	{
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		log.Debug("trying to connect")
		if err := b.c.Connect(ctx); err != nil {
			return err
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := b.c.Ping(ctx, nil); err != nil {
			return err
		}
	}

	return nil
}

func (b *Storage) Collection() *mongo.Collection {
	return b.c.Database(b.databaseName).Collection(b.collectionName)
}

func (b *Storage) Core() *mongo.Client {
	return b.c
}

func (b *Storage) Close() error {
	return b.c.Disconnect(context.Background())
}

func (b *Storage) Batch() storage.BatchStorage {
	return NewBatch(b)
}

func (b *Storage) Has(k string) (bool, error) {
	q := bson.M{"_k": resolveKey(k)}
	cur, err := b.Collection().Find(
		context.Background(),
		q,
		options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(1),
	)
	if err != nil {
		return false, err
	}

	return cur.Next(context.Background()), cur.Err()
}

func (b *Storage) MustExist(k string) error {
	exists, err := b.Has(k)
	if err != nil {
		return err
	} else if !exists {
		return storage.NotFound.New()
	}

	return nil
}

func (b *Storage) MustNotExist(k string) error {
	exists, err := b.Has(k)
	if err != nil {
		return err
	} else if exists {
		return storage.AlreadyExists.New()
	}

	return nil
}

func (b *Storage) Get(k string, v interface{}) error {
	if err := b.MustExist(k); err != nil {
		return err
	}

	r := b.Collection().FindOne(context.Background(), bson.M{"_k": resolveKey(k)})
	if err := r.Err(); err != nil {
		return err
	}

	raw, err := r.DecodeBytes()
	if err != nil {
		return err
	}

	_, err = UnmarshalDocument(raw, v)
	return err
}

func (b *Storage) Iterator(prefix string, v interface{}, opt storage.ListOptions) (func() (storage.Record, bool), func()) {
	q := bson.M{"_k": bson.M{"$regex": "^" + regexp.QuoteMeta(resolveKey(prefix))}}

	reverse := 1
	if opt.Reverse() {
		reverse = -1
	}

	if len(opt.Cursor()) > 0 {
		var dir string
		if opt.Reverse() {
			dir = "$lt"
		} else {
			dir = "$gt"
		}

		q = bson.M{
			"$and": bson.A{
				q,
				bson.M{"_k": bson.M{dir: resolveKey(string(opt.Cursor()))}},
			},
		}
	}

	mopt := options.Find().
		SetLimit(int64(opt.Limit())).
		SetSort(bson.M{"_k": reverse})

	cur, err := b.Collection().Find(context.Background(), q, mopt)

	return func() (storage.Record, bool) {
			// TODO err should be returned
			if err != nil {
				return storage.Record{}, false
			}

			hasNext := cur.Next(context.Background())
			if !hasNext {
				return storage.Record{}, false
			}

			nv := reflect.New(reflect.TypeOf(v)).Interface()
			doc, err := UnmarshalDocument([]byte(cur.Current), nv)
			if err != nil {
				return storage.Record{}, false
			}

			return storage.NewRecord(doc.K, reflect.ValueOf(nv).Elem().Interface()), true
		}, func() {
			if err != nil {
				return
			}

			cur.Close(context.Background())
		}
}

func (b *Storage) Insert(k string, v interface{}) error {
	if err := b.MustNotExist(k); err != nil {
		return err
	}

	doc, err := NewDocument(resolveKey(k), v)
	if err != nil {
		return err
	}
	_, err = b.Collection().InsertOne(context.Background(), doc)
	return err
}

func (b *Storage) Update(k string, v interface{}) error {
	if err := b.MustExist(k); err != nil {
		return err
	}

	if err := b.Delete(k); err != nil {
		return err
	}
	if err := b.Insert(k, v); err != nil {
		return err
	}
	return nil
}

func (b *Storage) Delete(k string) error {
	if err := b.MustExist(k); err != nil {
		return err
	}

	_, err := b.Collection().DeleteOne(context.Background(), bson.M{"_k": resolveKey(k)}, nil)
	return err
}

func (b *Storage) MultipleInsert(items ...storage.Value) error {
	return nil
}

func (b *Storage) MultipleUpdate(items ...storage.Value) error {
	return nil
}

func (b *Storage) MultipleDelete(keys ...string) error {
	return nil
}

func (b *Storage) Event(event string, values ...interface{}) {
	storage.Observer.Trigger(event, values...)
}

func resolveKey(key string) string {
	n := hex.EncodeToString([]byte(string(key[:2])))
	return string(n) + string(key[2:])
}
