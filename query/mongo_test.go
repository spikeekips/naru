package query

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"github.com/spikeekips/naru/common"
)

type baseTestMongo struct {
	suite.Suite
	client         *mongo.Client
	databaseName   string
	collectionName string
}

func (t *baseTestMongo) SetupTest() {
	client, err := connect()
	if err != nil {
		panic(client)
	}

	t.client = client
	t.databaseName = common.SequentialUUID()
	t.collectionName = common.SequentialUUID()
}

func (t *baseTestMongo) collection() *mongo.Collection {
	db := t.client.Database(t.databaseName, nil)
	return db.Collection(t.collectionName, nil)
}

func (t *baseTestMongo) TearDownTest() {
	if t.client == nil {
		return
	}
	db := t.client.Database(t.databaseName, nil)
	collection := db.Collection(t.collectionName, nil)

	if err := collection.Drop(nil); err != nil {
		log.Error("failed to drop collection", "error", err)
	} else {
		log.Debug("collection dropped", "collection", t.collectionName)
	}
	if err := db.Drop(nil); err != nil {
		log.Error("failed to drop db", "error", err)
	} else {
		log.Debug("db dropped", "db", t.databaseName)
	}

	disconnect(t.client)
}

type testMongo struct {
	baseTestMongo
}

func (t *testMongo) TestPing() {
	err := t.client.Ping(context.Background(), nil)
	t.NoError(err)
}

func (t *testMongo) TestInsert() {
	res, err := t.collection().InsertOne(context.Background(), bson.D{{"hello", "world"}})
	t.NoError(err)
	t.NotEmpty(res.InsertedID)
}

func (t *testMongo) TestInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world"})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "out there"})

	{
		count, err := t.collection().CountDocuments(context.Background(), bson.M{}, nil)
		t.NoError(err)
		t.Equal(int64(2), count)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), bson.D{{"hello", "world"}}, nil)
		t.NoError(err)
		t.Equal(int64(1), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), bson.M{"hello": "world"})
		defer cur.Close(context.Background())

		var count int
		for cur.Next(context.Background()) {
			var value map[string]interface{}
			err := cur.Decode(&value)
			t.NoError(err)

			t.False(value["_id"].(primitive.ObjectID).IsZero())
			t.NotEmpty(value["_id"].(primitive.ObjectID).String())
			t.Equal("world", value["hello"].(string))

			count += 1
		}
		err = cur.Err()
		t.NoError(err)
		t.Equal(1, count)
	}
}

func (t *testMongo) TestINInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	q := bson.M{
		"height": bson.M{
			"$in": [2]int{10, 1},
		},
	}
	{
		count, err := t.collection().CountDocuments(context.Background(), q, nil)
		t.NoError(err)
		t.Equal(int64(2), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), q)
		defer cur.Close(context.Background())

		var records []map[string]interface{}
		for cur.Next(context.Background()) {
			var value map[string]interface{}
			err := cur.Decode(&value)
			t.NoError(err)

			records = append(records, value)
		}
		err = cur.Err()
		t.NoError(err)

		t.Equal(2, len(records))
		t.Equal(int32(10), records[0]["height"].(int32))
		t.Equal(int32(1), records[1]["height"].(int32))
	}
}

func (t *testMongo) TestCheckExistsFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	{ // exists
		q := bson.M{"height": 10}
		cur, err := t.collection().Find(
			context.Background(),
			q,
			options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(1),
		)
		defer cur.Close(context.Background())

		t.True(cur.Next(context.Background()))
		err = cur.Err()
		t.NoError(err)
	}

	{ // does not exist
		q := bson.M{"height": 100}
		cur, err := t.collection().Find(
			context.Background(),
			q,
			options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(1),
		)
		defer cur.Close(context.Background())

		t.False(cur.Next(context.Background()))
		err = cur.Err()
		t.NoError(err)
	}
}

func (t *testMongo) TestCheckExistsFindContext() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
		}
	}()

	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	{ // context.WithTimeout works in find
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
		defer cancel()

		q := bson.M{"height": 10}
		_, err := t.collection().Find(ctx, q, nil)
		t.True(strings.Contains(err.Error(), "context deadline exceeded"))
	}

	{ // context.WithTimeout in cursor does not work
		q := bson.M{"height": 10}
		cur, err := t.collection().Find(
			context.Background(),
			q,
			options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(1),
		)
		t.NoError(err)
		defer cur.Close(context.Background())

		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
		defer cancel()

		t.True(cur.Next(ctx))
		t.NotEmpty(cur.Current)

		var value map[string]interface{}
		err = cur.Decode(&value)
		t.NoError(err)
		t.NotEmpty(value)
	}
}

func (t *testMongo) TestCheckExistsFindOneContext() {
	defer func() {
		t.NotEmpty(recover())
	}()

	{ // context.WithTimeout works in findOne
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
		defer cancel()

		q := bson.M{"height": 10}
		_ = t.collection().FindOne(ctx, q, nil)
	}
}

func (t *testMongo) TestCheckExistsFindOne() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	{
		q := bson.M{"height": 10}
		r := t.collection().FindOne(
			context.Background(),
			q,
			options.FindOne().SetProjection(bson.M{"_id": 1}),
		)
		raw, err := r.DecodeBytes()
		t.NoError(err)
		v, err := raw.Values()
		t.NoError(err)
		t.NotEmpty(v)
		el, err := raw.Elements()
		t.NoError(err)
		t.NotEmpty(el)
	}

	{
		q := bson.M{"height": 100}
		r := t.collection().FindOne(
			context.Background(),
			q,
			options.FindOne().SetProjection(bson.M{"_id": 1}),
		)
		raw, err := r.DecodeBytes()
		t.Equal(mongo.ErrNoDocuments, err)
		v, err := raw.Values()
		t.True(bsoncore.InsufficientBytesError{}.Equal(err))
		t.Equal(0, len(v))
		el, err := raw.Elements()
		t.True(bsoncore.InsufficientBytesError{}.Equal(err))
		t.Equal(0, len(el))
	}
}

func TestMongo(t *testing.T) {
	if client, err := connect(); err != nil {
		log.Warn("mongodb test will be skipped")
		return
	} else {
		defer disconnect(client)
	}

	suite.Run(t, new(testMongo))
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
		log.Debug("trying to connect")
		err = client.Connect(ctx)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("connected")

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
	log.Debug("disconnected", "client", client_id(client))
	if client == nil {
		return
	}

	client.Disconnect(context.Background())
}
