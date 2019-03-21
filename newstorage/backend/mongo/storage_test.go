package mongostorage

import (
	"context"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/config"
	"github.com/spikeekips/naru/newstorage"
	"github.com/spikeekips/naru/newstorage/item"
)

type baseTestMongoStorage struct {
	suite.Suite
	s            *Storage
	databaseName string
}

func (t *baseTestMongoStorage) SetupTest() {
	t.databaseName = common.SequentialUUID()

	log.Debug("trying to connect", "db", t.databaseName)
	c := &config.MongoStorage{
		URI: mongooptions.Client().ApplyURI("mongodb://localhost:27017"),
		DB:  t.databaseName,
	}

	s, err := NewStorage(c)
	t.NoError(err)
	t.s = s
	log.Debug("connected", "db", t.databaseName)
}

func (t *baseTestMongoStorage) TearDownTest() {
	if t.s == nil {
		return
	}
	db := t.s.Core().Database(t.databaseName, nil)
	collection := db.Collection(defaultCollectionName, nil)

	if err := collection.Drop(nil); err != nil {
		log.Error("failed to drop collection", "error", err)
	} else {
		log.Debug("collection dropped", "collection", defaultCollectionName)
	}
	if err := db.Drop(nil); err != nil {
		log.Error("failed to drop db", "error", err)
	} else {
		log.Debug("db dropped", "db", t.databaseName)
	}

	log.Debug("disconnected", "db", t.databaseName)
	err := t.s.Close()
	t.NoError(err)
}

type testMongoStorage struct {
	baseTestMongoStorage
}

type testMongoStorageItem struct {
	A    string   `bson:"a"`
	B    int      `bson:"b"`
	BOld int      `bson:"bold"`
	C    []uint64 `bson:"c"`
}

func (t *testMongoStorage) TestKey() {
	key := "showme"
	o := resolveKey(item.InternalPrefix + key)
	t.Equal("0000showme", o)

	n := resolveKey(item.InternalPrefix + "showme")
	t.Equal(o, n)
}

func (t *testMongoStorage) TestInsert() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r)
		}
	}()

	var items []testMongoStorageItem
	for i := uint64(0); i < 5; i++ {
		item := testMongoStorageItem{
			A: "\x00\x10-" + common.SequentialUUID(),
			B: int(i),
			C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
		}
		err := t.s.Insert(item.A, item)
		t.NoError(err)

		items = append(items, item)
	}

	{ // other records
		for i := uint64(0); i < 3; i++ {
			item := testMongoStorageItem{
				A: "BBB-" + common.SequentialUUID(),
				B: int(i),
				C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
			}
			err := t.s.Insert(item.A, item)
			t.NoError(err)
		}
	}

	iter, cls := t.s.Iterator("\x00\x10-", testMongoStorageItem{}, newstorage.NewDefaultListOptions(false, nil, 0))

	var records []newstorage.Record
	for {
		record, next := iter()
		if !next {
			break
		}

		records = append(records, record)
	}
	cls()

	t.Equal(len(items), len(records))

	for n, item := range items {
		found, err := t.s.Has(item.A)
		t.NoError(err)
		t.True(found)

		t.IsType(testMongoStorageItem{}, records[n].Value)
		t.Equal(item.A, records[n].Value.(testMongoStorageItem).A)
	}
}

func (t *testMongoStorage) TestInsertBatch() {
	batch := t.s.Batch()

	var items []testMongoStorageItem
	for i := uint64(0); i < 5; i++ {
		item := testMongoStorageItem{
			A: "AAA-" + common.RandomUUID(),
			B: int(i),
			C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
		}
		err := batch.Insert(item.A, item)
		t.NoError(err)

		items = append(items, item)
	}

	// before `Batch.Write()`
	for _, item := range items {
		found, err := t.s.Has(item.A)
		t.NoError(err)
		t.False(found)
	}

	err := batch.Write()
	t.NoError(err)

	for _, item := range items {
		found, err := t.s.Has(item.A)
		t.NoError(err)
		t.True(found)
	}
}

func (t *testMongoStorage) TestUpdateBatch() {
	var items []testMongoStorageItem
	for i := uint64(0); i < 5; i++ {
		item := testMongoStorageItem{
			A:    "AAA-" + common.RandomUUID(),
			B:    int(i),
			BOld: int(i),
			C:    []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
		}
		err := t.s.Insert(item.A, item)
		t.NoError(err)

		items = append(items, item)
	}

	batch := t.s.Batch()
	for _, item := range items {
		item.B = len(items) + item.B
		err := batch.Update(item.A, item)
		t.NoError(err)
	}

	for _, item := range items {
		var ni testMongoStorageItem
		err := t.s.Get(item.A, &ni)
		t.NoError(err)
		t.Equal(item.B, item.BOld)
		t.Equal(item.B, ni.B)
		t.Equal(item.BOld, ni.BOld)
	}

	err := batch.Write()
	t.NoError(err)

	for _, item := range items {
		var ni testMongoStorageItem
		err := t.s.Get(item.A, &ni)
		t.NoError(err)
		t.Equal(len(items)+item.B, ni.B)
	}
}

func (t *testMongoStorage) TestDeleteBatch() {
	var items []testMongoStorageItem
	for i := uint64(0); i < 5; i++ {
		item := testMongoStorageItem{
			A:    "AAA-" + common.RandomUUID(),
			B:    int(i),
			BOld: int(i),
			C:    []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
		}
		err := t.s.Insert(item.A, item)
		t.NoError(err)

		items = append(items, item)
	}

	batch := t.s.Batch()
	for _, item := range items {
		item.B = len(items) + item.B
		err := batch.Delete(item.A)
		t.NoError(err)
	}

	for _, item := range items {
		found, err := t.s.Has(item.A)
		t.NoError(err)
		t.True(found)
	}

	err := batch.Write()
	t.NoError(err)

	for _, item := range items {
		found, err := t.s.Has(item.A)
		t.NoError(err)
		t.False(found)
	}
}

func (t *testMongoStorage) TestIteratorLimit() {
	prefix := "\x00\x10-=---"

	var items []testMongoStorageItem
	for i := uint64(0); i < 5; i++ {
		item := testMongoStorageItem{
			A: prefix + common.SequentialUUID(),
			B: int(i),
			C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
		}
		err := t.s.Insert(item.A, item)
		t.NoError(err)

		items = append(items, item)
	}

	{ // other records
		for i := uint64(0); i < 3; i++ {
			item := testMongoStorageItem{
				A: "BBB-" + common.SequentialUUID(),
				B: int(i),
				C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
			}
			err := t.s.Insert(item.A, item)
			t.NoError(err)
		}
	}

	{ // reverse=false, no cursor, 3 limited
		limit := 3
		opt := newstorage.NewDefaultListOptions(false, nil, uint64(limit))
		iter, cls := t.s.Iterator(prefix, testMongoStorageItem{}, opt)

		var records []newstorage.Record
		for {
			record, next := iter()
			if !next {
				break
			}

			records = append(records, record)
		}
		cls()

		t.Equal(limit, len(records))

		for n, item := range items[:limit] {
			t.Equal(item.A, records[n].Value.(testMongoStorageItem).A)
			t.Equal(item.B, records[n].Value.(testMongoStorageItem).B)
			for m, i := range item.C {
				t.Equal(i, records[n].Value.(testMongoStorageItem).C[m])
			}
		}
	}

	{ // reverse=true, no cursor, 3 limited
		limit := 3
		opt := newstorage.NewDefaultListOptions(true, nil, uint64(limit))
		iter, cls := t.s.Iterator(prefix, testMongoStorageItem{}, opt)

		var records []newstorage.Record
		for {
			record, next := iter()
			if !next {
				break
			}

			records = append(records, record)
		}
		cls()

		t.Equal(limit, len(records))

		for n, item := range items[len(items)-limit : limit] {
			t.Equal(item.A, records[limit-n-1].Value.(testMongoStorageItem).A)
			t.Equal(item.B, records[limit-n-1].Value.(testMongoStorageItem).B)
			for m, i := range item.C {
				t.Equal(i, records[limit-n-1].Value.(testMongoStorageItem).C[m])
			}
		}
	}

	{ // reverse=true, 4th cursor, 10 limited
		cursorItem := items[4]
		limit := 10
		opt := newstorage.NewDefaultListOptions(true, []byte(cursorItem.A), uint64(limit))
		iter, cls := t.s.Iterator(prefix, testMongoStorageItem{}, opt)

		var records []newstorage.Record
		for {
			record, next := iter()
			if !next {
				break
			}

			records = append(records, record)
		}
		cls()

		t.Equal(4, len(records))

		for n, item := range items[:4] {
			t.Equal(item.A, records[3-n].Value.(testMongoStorageItem).A)
			t.Equal(item.B, records[3-n].Value.(testMongoStorageItem).B)
			for m, i := range item.C {
				t.Equal(i, records[3-n].Value.(testMongoStorageItem).C[m])
			}
		}
	}
}

func (t *testMongoStorage) TestIteratorCursor() {
	prefix := "\x00\x10-=---"

	var items []testMongoStorageItem
	for i := uint64(0); i < 5; i++ {
		item := testMongoStorageItem{
			A: prefix + common.SequentialUUID(),
			B: int(i),
			C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
		}
		err := t.s.Insert(item.A, item)
		t.NoError(err)

		items = append(items, item)
	}

	{ // other records
		for i := uint64(0); i < 3; i++ {
			item := testMongoStorageItem{
				A: "BBB-" + common.SequentialUUID(),
				B: int(i),
				C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
			}
			err := t.s.Insert(item.A, item)
			t.NoError(err)
		}
	}

	{ // reverse=false, 3rd cursor, unlimited
		cursorItem := items[2]
		opt := newstorage.NewDefaultListOptions(false, []byte(cursorItem.A), 0)
		iter, cls := t.s.Iterator(prefix, testMongoStorageItem{}, opt)

		var records []newstorage.Record
		for {
			record, next := iter()
			if !next {
				break
			}

			records = append(records, record)
		}
		cls()

		t.Equal(2, len(records))

		for n, item := range items[3:] {
			t.Equal(item.A, records[n].Value.(testMongoStorageItem).A)
			t.Equal(item.B, records[n].Value.(testMongoStorageItem).B)
			for m, i := range item.C {
				t.Equal(i, records[n].Value.(testMongoStorageItem).C[m])
			}
		}
	}

	{ // reverse=true, 3rd cursor, unlimited
		cursorItem := items[2]
		opt := newstorage.NewDefaultListOptions(true, []byte(cursorItem.A), 0)
		iter, cls := t.s.Iterator(prefix, testMongoStorageItem{}, opt)

		var records []newstorage.Record
		for {
			record, next := iter()
			if !next {
				break
			}

			records = append(records, record)
		}
		cls()

		t.Equal(2, len(records))

		for n, item := range items[:2] {
			t.Equal(item.A, records[1-n].Value.(testMongoStorageItem).A)
			t.Equal(item.B, records[1-n].Value.(testMongoStorageItem).B)
			for m, i := range item.C {
				t.Equal(i, records[1-n].Value.(testMongoStorageItem).C[m])
			}
		}
	}
}

func (t *testMongoStorage) TestIteratorOptions() {
	prefix := "\x00\x10-=---"

	var items []testMongoStorageItem
	for i := uint64(0); i < 5; i++ {
		item := testMongoStorageItem{
			A: prefix + common.SequentialUUID(),
			B: int(i),
			C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
		}
		err := t.s.Insert(item.A, item)
		t.NoError(err)

		items = append(items, item)
	}

	{ // other records
		for i := uint64(0); i < 3; i++ {
			item := testMongoStorageItem{
				A: "BBB-" + common.SequentialUUID(),
				B: int(i),
				C: []uint64{(i * 3), (i * 3) + 1, (i * 3) + 2},
			}
			err := t.s.Insert(item.A, item)
			t.NoError(err)
		}
	}

	{ // reverse=false, no cursor, unlimited
		opt := newstorage.NewDefaultListOptions(false, nil, 0)
		iter, cls := t.s.Iterator(prefix, testMongoStorageItem{}, opt)

		var records []newstorage.Record
		for {
			record, next := iter()
			if !next {
				break
			}

			records = append(records, record)
		}
		cls()

		t.Equal(len(items), len(records))

		for n, item := range items {
			t.Equal(item.A, records[n].Value.(testMongoStorageItem).A)
			t.Equal(item.B, records[n].Value.(testMongoStorageItem).B)
			for m, i := range item.C {
				t.Equal(i, records[n].Value.(testMongoStorageItem).C[m])
			}
		}
	}

	{ // reverse=true, no cursor, unlimited
		opt := newstorage.NewDefaultListOptions(true, nil, 0)
		iter, cls := t.s.Iterator(prefix, testMongoStorageItem{}, opt)

		var records []newstorage.Record
		for {
			record, next := iter()
			if !next {
				break
			}

			records = append(records, record)
		}
		cls()

		t.Equal(len(items), len(records))

		for n, item := range items {
			t.Equal(item.A, records[len(items)-n-1].Value.(testMongoStorageItem).A)
			t.Equal(item.B, records[len(items)-n-1].Value.(testMongoStorageItem).B)
			for m, i := range item.C {
				t.Equal(i, records[len(items)-n-1].Value.(testMongoStorageItem).C[m])
			}
		}
	}
}

func TestMongoStorage(t *testing.T) {
	if client, err := connect(); err != nil {
		log.Warn("mongodb test will be skipped")
		return
	} else {
		disconnect(client)
	}

	suite.Run(t, new(testMongoStorage))
}

func connect() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	connectTimeout := time.Second * 1
	socketTimeout := time.Second * 1
	clientOptions.ConnectTimeout = &connectTimeout
	clientOptions.SocketTimeout = &socketTimeout

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err = client.Connect(ctx)
		if err != nil {
			return nil, err
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := client.Ping(ctx, nil); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func client_id(client *mongo.Client) string {
	return fmt.Sprintf("%p", client)
}

func disconnect(client *mongo.Client) {
	if client == nil {
		return
	}

	client.Disconnect(context.Background())
}
