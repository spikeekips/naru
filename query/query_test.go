package query

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
)

type testTermQuery struct {
	suite.Suite
}

func (t *testTermQuery) TestNew() {
	tm, _ := NewTerm("this-field", "this-value")
	tq := NewTermQuery(IS, tm)

	_, ok := (interface{})(tq).(Query)
	t.True(ok)
	t.Equal(IS, tq.Operator())
	t.Equal(tm.Field(), tq.Term().Field())
	t.True(tm.Equal(tq.Term()))
}

func (t *testTermQuery) TestConjunect() {
	var a, b TermQuery
	{
		tm, _ := NewTerm("this-a", "this-a")
		a = NewTermQuery(IS, tm)
	}
	{
		tm, _ := NewTerm("this-b", 2)
		b = NewTermQuery(IS, tm)
	}

	c := a.Conjunct(OR, b)
	t.Equal(OR, c.Conjunction())
	t.True(a.Equal(c.Queries()[0]))
	t.True(b.Equal(c.Queries()[1]))
}

func (t *testTermQuery) TestAppend() {
	var a, b TermQuery
	{
		tm, _ := NewTerm("this-a", "this-a")
		a = NewTermQuery(IS, tm)
	}
	{
		tm, _ := NewTerm("this-b", 2)
		b = NewTermQuery(IS, tm)
	}

	c := a.Append(b)
	t.Equal(AND, c.Conjunction())
	t.True(a.Equal(c.Queries()[0]))
	t.True(b.Equal(c.Queries()[1]))
}

func (t *testTermQuery) TestIterator() {
	tm, _ := NewTerm("this-a", "this-a")
	a := NewTermQuery(IS, tm)

	var queries []Query
	for q := range a.Iterator() {
		queries = append(queries, q)
	}

	t.Equal(1, len(queries))
	t.True(a.Equal(queries[0]))
}

type testConjunctionQuery struct {
	suite.Suite
}

func (t *testConjunctionQuery) makeQuery(conjunction Conjunction, a, b Query) ConjunctionQuery {
	return NewConjunctionQuery(conjunction, a, b)
}

func (t *testConjunctionQuery) TestNew() {
	var a, b TermQuery
	{
		tm, _ := NewTerm("this-a", "this-a")
		a = NewTermQuery(IS, tm)
	}
	{
		tm, _ := NewTerm("this-b", 2)
		b = NewTermQuery(IS, tm)
	}

	q := t.makeQuery(OR, a, b)
	t.Equal(OR, q.Conjunction())
	t.True(a.Equal(q.Queries()[0]))
	t.True(b.Equal(q.Queries()[1]))
}

func (t *testConjunctionQuery) TestNewFromConjunctionQuery() {
	var a TermQuery
	var b ConjunctionQuery
	{
		tm, _ := NewTerm("this-a", "this-a")
		a = NewTermQuery(IS, tm)
	}
	{
		tm, _ := NewTerm("this-c", 2)
		c := NewTermQuery(IS, tm)
		b = NewConjunctionQuery(AND, a, c)
	}

	q := t.makeQuery(AND, a, b)
	t.Equal(ConjunctionQueryType, q.Type())
	t.Equal(AND, q.Conjunction())
	t.True(a.Equal(q.Queries()[0]))
	t.True(b.Equal(q.Queries()[1]))
}

func (t *testConjunctionQuery) TestConjunct() {
	var a TermQuery
	var b ConjunctionQuery
	{
		tm, _ := NewTerm("this-a", "this-a")
		a = NewTermQuery(IS, tm)
	}
	{
		tm, _ := NewTerm("this-c", 2)
		c := NewTermQuery(IS, tm)
		b = NewConjunctionQuery(AND, a, c)
	}

	q := b.Conjunct(AND, a)
	t.Equal(ConjunctionQueryType, q.Type())
	t.Equal(AND, q.Conjunction())
	t.True(b.Equal(q.Queries()[0]))
	t.True(a.Equal(q.Queries()[1]))
}

func (t *testConjunctionQuery) TestAppend() {
	var a, b TermQuery
	var c ConjunctionQuery
	{
		tm, _ := NewTerm("this-a", "this-a")
		a = NewTermQuery(IS, tm)
	}
	{
		tm, _ := NewTerm("this-c", 2)
		b = NewTermQuery(IS, tm)
		c = NewConjunctionQuery(AND, b)
	}

	q := c.Append(a)
	t.Equal(ConjunctionQueryType, q.Type())
	t.Equal(c.Conjunction(), q.Conjunction())
	t.True(b.Equal(q.Queries()[0]))
	t.True(a.Equal(q.Queries()[1]))
}

func (t *testConjunctionQuery) TestIterator() {
	var a, b TermQuery
	var c ConjunctionQuery
	{
		tm, _ := NewTerm("this-a", "this-a")
		a = NewTermQuery(IS, tm)
	}
	{
		tm, _ := NewTerm("this-c", 2)
		b = NewTermQuery(IS, tm)
		c = NewConjunctionQuery(AND, b)
	}

	d := c.Append(a)

	var queries []Query
	for q := range d.Iterator() {
		queries = append(queries, q)
	}
	t.Equal(3, len(queries))
	t.True(d.Equal(queries[0]))
	t.True(b.Equal(queries[1]))
	t.True(a.Equal(queries[2]))
}

type testQueryBuilder struct {
	suite.Suite
}

func (t *testQueryBuilder) TestNew() {
	var a, b TermQuery
	var c ConjunctionQuery
	{
		tm, _ := NewTerm("this-a", "this-a-value")
		a = NewTermQuery(NOT, tm)
	}
	{
		tm, _ := NewTerm("this-c", 33)
		b = NewTermQuery(IS, tm)
		c = NewConjunctionQuery(AND, b)
	}

	d := c.Append(a)

	builder := TestMongoBuilder{query: d}
	mongoQuery, err := builder.Build()
	t.NoError(err)
	t.IsType(bson.D{}, mongoQuery)

	{ // bson.Marshal
		j, err := bson.Marshal(mongoQuery.(bson.D))
		t.NoError(err)
		t.Equal([]byte{89, 0, 0, 0, 4, 36, 97, 110, 100, 0, 78, 0, 0, 0, 3, 48, 0, 27, 0, 0, 0, 3, 116, 104, 105, 115, 45, 99, 0, 14, 0, 0, 0, 16, 36, 101, 113, 0, 33, 0, 0, 0, 0, 0, 3, 49, 0, 40, 0, 0, 0, 3, 116, 104, 105, 115, 45, 97, 0, 27, 0, 0, 0, 2, 36, 110, 101, 0, 13, 0, 0, 0, 116, 104, 105, 115, 45, 97, 45, 118, 97, 108, 117, 101, 0, 0, 0, 0, 0}, j)
	}

	{ // bson.MarshalExtJSON
		j, err := bson.MarshalExtJSON(mongoQuery.(bson.D), false, false)
		t.NoError(err)
		t.Equal(`{"$and":[{"this-c":{"$eq":33}},{"this-a":{"$ne":"this-a-value"}}]}`, string(j))
	}
}

func TestTermQuery(t *testing.T) {
	suite.Run(t, new(testTermQuery))
}

func TestConjunctionQuery(t *testing.T) {
	suite.Run(t, new(testConjunctionQuery))
}

func TestQueryBuilder(t *testing.T) {
	suite.Run(t, new(testQueryBuilder))
}
