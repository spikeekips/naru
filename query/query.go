package query

import (
	"encoding/json"
)

type QueryType uint

const (
	InvalidQuery QueryType = iota
	TermQueryType
	ConjunctionQueryType
)

func (q QueryType) MarshalJSON() ([]byte, error) {
	if q == InvalidQuery || q > ConjunctionQueryType {
		return nil, InvalidQueryType.New()
	}

	return json.Marshal(q.String())
}

func (q QueryType) String() string {
	switch q {
	case TermQueryType:
		return "TermQuery"
	case ConjunctionQueryType:
		return "ConjunctionQuery"
	}

	return "InvalidQuery"
}

type Query interface {
	String() string
	Type() QueryType
	Equal(Query) bool
	Conjunct(Conjunction, ...Query) ConjunctionQuery
	Append(Query) ConjunctionQuery
	Iterator() <-chan Query
	Queries() []Query
}

type TermQuery struct {
	operator Operator
	term     Term
}

func NewTermQuery(operator Operator, term Term) TermQuery {
	return TermQuery{operator: operator, term: term}
}

func (t TermQuery) Type() QueryType {
	return TermQueryType
}

func (t TermQuery) MarshalJSON() ([]byte, error) {
	j, _ := json.Marshal(t.term)
	b, err := json.Marshal(map[string]interface{}{
		"type":     t.Type(),
		"operator": t.operator.String(),
		"term":     json.RawMessage(j),
	})

	return b, err
}

func (t TermQuery) Equal(v Query) bool {
	if v.Type() != TermQueryType {
		return false
	}

	tq := v.(TermQuery)
	if t.operator != tq.Operator() {
		return false
	}

	if !t.term.Equal(tq.Term()) {
		return false
	}

	return true
}

func (t TermQuery) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

func (t TermQuery) Conjunct(conjunction Conjunction, queries ...Query) ConjunctionQuery {
	nq := []Query{t}
	return NewConjunctionQuery(conjunction, append(nq, queries...)...)
}

func (t TermQuery) Append(q Query) ConjunctionQuery {
	return NewConjunctionQuery(AND, t, q)
}

func (t TermQuery) Operator() Operator {
	return t.operator
}

func (t TermQuery) Term() Term {
	return t.term
}

func (t TermQuery) Queries() []Query {
	return nil
}

func (t TermQuery) Iterator() <-chan Query {
	c := make(chan Query, 10)

	go func() {
		c <- t
		close(c)
	}()

	return c
}

type ConjunctionQuery struct {
	conjunction Conjunction
	queries     []Query
}

func NewConjunctionQuery(conjunction Conjunction, queries ...Query) ConjunctionQuery {
	return ConjunctionQuery{conjunction: conjunction, queries: queries}
}

func (t ConjunctionQuery) Type() QueryType {
	return ConjunctionQueryType
}

func (t ConjunctionQuery) MarshalJSON() ([]byte, error) {
	var j []json.RawMessage
	for _, q := range t.queries {
		b, err := json.Marshal(q)
		if err != nil {
			return nil, err
		}

		j = append(j, json.RawMessage(b))
	}

	return json.Marshal(map[string]interface{}{
		"type":        t.Type(),
		"conjunction": t.conjunction.String(),
		"queries":     j,
	})
}

func (t ConjunctionQuery) String() string {
	b, _ := json.Marshal(t)
	return string(b)
}

func (t ConjunctionQuery) Conjunct(conjunction Conjunction, queries ...Query) ConjunctionQuery {
	nq := []Query{t}
	return NewConjunctionQuery(conjunction, append(nq, queries...)...)
}

func (t ConjunctionQuery) Append(q Query) ConjunctionQuery {
	n := make([]Query, len(t.queries))
	copy(n, t.queries)
	n = append(n, q)
	return NewConjunctionQuery(t.conjunction, n...)
}

func (t ConjunctionQuery) Conjunction() Conjunction {
	return t.conjunction
}

func (t ConjunctionQuery) Queries() []Query {
	return t.queries
}

func (t ConjunctionQuery) Equal(v Query) bool {
	if v.Type() != ConjunctionQueryType {
		return false
	}

	tq := v.(ConjunctionQuery)
	if t.conjunction != tq.Conjunction() {
		return false
	}

	if len(t.queries) != len(tq.Queries()) {
		return false
	}

	queries := tq.Queries()
	for i, q := range queries {
		if !q.Equal(queries[i]) {
			return false
		}
	}

	return true
}

func (t ConjunctionQuery) Iterator() <-chan Query {
	c := make(chan Query, 10)
	go func() {
		c <- t
		for _, q := range t.Queries() {
			for cq := range q.Iterator() {
				c <- cq
			}
		}
		close(c)
	}()

	return c
}
