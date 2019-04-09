package mongostorage

import (
	"context"
	"reflect"
	"regexp"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/storage"
)

var defaultCollectionName string = "default"
var defaultConnectTimeout = time.Second * 10

type Storage struct {
	sync.RWMutex
	c              *mongo.Client
	databaseName   string
	collectionName string
	config         config.MongoStorage
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
		config:         *c,
	}
	if err := s.connect(); err != nil {
		return nil, err
	}

	return s, nil
}

func (b *Storage) connect() error {
	{
		ctx, _ := context.WithTimeout(context.Background(), defaultConnectTimeout)
		if err := b.c.Connect(ctx); err != nil {
			return err
		}
	}

	{
		ctx, _ := context.WithTimeout(context.Background(), defaultConnectTimeout)
		if err := b.c.Ping(ctx, nil); err != nil {
			b.c.Disconnect(context.Background())
			return err
		}
	}

	return nil
}

func (b *Storage) New() (*Storage, error) {
	return NewStorage(&b.config)
}

func (b *Storage) Database() *mongo.Database {
	return b.c.Database(b.databaseName)
}

func (b *Storage) Collection(key string) (*mongo.Collection, error) {
	c, err := GetCollection(key)
	if err != nil {
		return nil, err
	}

	return b.Database().Collection(c), nil
}

func (b *Storage) Core() *mongo.Client {
	return b.c
}

func (b *Storage) Close() error {
	return b.c.Disconnect(context.Background())
}

func (b *Storage) Initialize() error {
	return b.Database().Drop(context.Background())
}

func (b *Storage) Batch() (storage.BatchStorage, error) {
	return NewBatch(b, true)
}

func (b *Storage) Has(k string) (bool, error) {
	q := bson.M{KEY: k}

	col, err := b.Collection(k)
	if err != nil {
		return false, err
	}

	cur, err := col.Find(
		context.Background(),
		q,
		options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(1),
	)
	if err != nil {
		return false, err
	}
	defer cur.Close(context.Background())

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

	col, err := b.Collection(k)
	if err != nil {
		return err
	}

	r := col.FindOne(context.Background(), bson.M{KEY: k})
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
	q := bson.M{KEY: bson.M{"$regex": "^" + regexp.QuoteMeta(prefix)}}

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
				bson.M{KEY: bson.M{dir: string(opt.Cursor())}},
			},
		}
	}

	mopt := options.Find().
		SetLimit(int64(opt.Limit())).
		SetSort(bson.M{KEY: reverse})

	col, err := b.Collection(prefix)
	if err != nil {
		return nil, nil
	}

	cur, err := col.Find(context.Background(), q, mopt)
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

	doc, err := NewDocument(k, v)
	if err != nil {
		return err
	}
	col, err := b.Collection(k)
	if err != nil {
		return err
	}

	_, err = col.InsertOne(context.Background(), doc)
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

	col, err := b.Collection(k)
	if err != nil {
		return err
	}

	_, err = col.DeleteOne(context.Background(), bson.M{KEY: k}, nil)
	return err
}

func (b *Storage) Event(event string, values ...interface{}) {
	storage.Observer.Trigger(event, values...)
}
