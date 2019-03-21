package query

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type testMongoQuery struct {
	baseTestMongo
}

func (t *testMongoQuery) TestSimpleInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world"})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "out there"})

	{
		count, err := t.collection().CountDocuments(context.Background(), bson.M{}, nil)
		t.NoError(err)
		t.Equal(int64(2), count)
	}

	bq := bson.M{"hello": bson.M{"$eq": "world"}}
	{
		count, err := t.collection().CountDocuments(context.Background(), bq, nil)
		t.NoError(err)
		t.Equal(int64(1), count)
	}

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq := NewTermQuery(IS, tm)

		builder := NewTestMongoBuilder(tq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(1), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), bq)
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

	{
		cur, err := t.collection().Find(context.Background(), mq)
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

func (t *testMongoQuery) TestORInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world"})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "out there"})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq0 := NewTermQuery(IS, tm)

		tm, _ = NewTerm("hello", "out there")
		tq1 := NewTermQuery(IS, tm)
		cq := tq0.Conjunct(OR, tq1)

		builder := NewTestMongoBuilder(cq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(2), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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

		t.Equal("world", records[0]["hello"].(string))
		t.Equal("out there", records[1]["hello"].(string))
	}
}

func (t *testMongoQuery) TestANDInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "people": "man"})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "people": "woman"})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "out there"})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq := NewTermQuery(IS, tm)

		builder := NewTestMongoBuilder(tq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(2), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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

		t.Equal("world", records[0]["hello"].(string))
		t.Equal("man", records[0]["people"].(string))
		t.Equal("world", records[1]["hello"].(string))
		t.Equal("woman", records[1]["people"].(string))
	}

	{
		tm, _ := NewTerm("hello", "world")
		tq0 := NewTermQuery(IS, tm)
		tm, _ = NewTerm("people", "man")
		tq1 := NewTermQuery(IS, tm)

		cq := tq0.Conjunct(AND, tq1)

		builder := NewTestMongoBuilder(cq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(1), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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
		t.Equal(1, len(records))

		t.Equal("world", records[0]["hello"].(string))
		t.Equal("man", records[0]["people"].(string))
	}
}

func (t *testMongoQuery) TestGTInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq0 := NewTermQuery(IS, tm)

		tm, _ = NewTerm("height", 9)
		tq1 := NewTermQuery(GT, tm)
		cq := tq0.Conjunct(AND, tq1)

		builder := NewTestMongoBuilder(cq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(1), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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
		t.Equal(1, len(records))

		t.Equal("world", records[0]["hello"].(string))

		// basically bson treats int values under 64 as `int32`
		t.Equal(int32(10), records[0]["height"].(int32))
	}
}

func (t *testMongoQuery) TestINInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq0 := NewTermQuery(IS, tm)

		tm, _ = NewTerm("height", []int{10, 1})
		tq1 := NewTermQuery(IN, tm)
		cq := tq0.Conjunct(AND, tq1)

		builder := NewTestMongoBuilder(cq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(2), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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

		t.Equal("world", records[0]["hello"].(string))
		t.Equal(int32(10), records[0]["height"].(int32))
		t.Equal("world", records[1]["hello"].(string))
		t.Equal(int32(1), records[1]["height"].(int32))
	}
}

func (t *testMongoQuery) TestNEInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq0 := NewTermQuery(IS, tm)

		tm, _ = NewTerm("height", 9)
		tq1 := NewTermQuery(NOT, tm)
		cq := tq0.Conjunct(AND, tq1)

		builder := NewTestMongoBuilder(cq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(2), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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

		t.Equal("world", records[0]["hello"].(string))
		t.Equal(int32(10), records[0]["height"].(int32))
		t.Equal("world", records[1]["hello"].(string))
		t.Equal(int32(1), records[1]["height"].(int32))
	}
}

func (t *testMongoQuery) TestNOTINInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq0 := NewTermQuery(IS, tm)

		tm, _ = NewTerm("height", []int{10, 1})
		tq1 := NewTermQuery(NOTIN, tm)
		cq := tq0.Conjunct(AND, tq1)

		builder := NewTestMongoBuilder(cq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(1), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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
		t.Equal(1, len(records))

		t.Equal("world", records[0]["hello"].(string))
		t.Equal(int32(9), records[0]["height"].(int32))
	}
}

func (t *testMongoQuery) TestFloatInsertAndFind() {
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 10})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 9})
	t.collection().InsertOne(context.Background(), bson.M{"hello": "world", "height": 1})
	t.collection().InsertOne(context.Background(), bson.M{"world": "hello"})

	var mq bson.D
	{
		tm, _ := NewTerm("hello", "world")
		tq0 := NewTermQuery(IS, tm)

		tm, _ = NewTerm("height", 9.1)
		tq1 := NewTermQuery(GT, tm)
		cq := tq0.Conjunct(AND, tq1)

		builder := NewTestMongoBuilder(cq)
		q, err := builder.Build()
		t.NoError(err)

		mq = q.(bson.D)
	}

	{
		count, err := t.collection().CountDocuments(context.Background(), mq, nil)
		t.NoError(err)
		t.Equal(int64(1), count)
	}

	{
		cur, err := t.collection().Find(context.Background(), mq)
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
		t.Equal(1, len(records))

		t.Equal("world", records[0]["hello"].(string))
		t.Equal(int32(10), records[0]["height"].(int32))
	}
}

func TestMongoQuery(t *testing.T) {
	if client, err := connect(); err != nil {
		log.Warn("mongodb test will be skipped")
		return
	} else {
		defer disconnect(client)
	}

	suite.Run(t, new(testMongoQuery))
}
