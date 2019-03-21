package query

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

var mongoComparisonByOperator = map[Operator]string{
	IS:  "$eq",
	NOT: "$ne",
	GT:  "$gt",
	GTE: "$gte",
	LT:  "$lt",
	LTE: "$lte",
	IN:  "$in",
}

var mongoLogicalByConjunction = map[Conjunction]string{
	AND: "$and",
	OR:  "$or",
}

func mongoOperator(operator Operator) (string, error) {
	m, found := mongoComparisonByOperator[operator]
	if !found {
		return "", NotSupportedOperator.New()
	}

	return m, nil
}

func mongoConjunction(conjunction Conjunction) (string, error) {
	m, found := mongoLogicalByConjunction[conjunction]
	if !found {
		return "", NotSupportedConjunction.New()
	}

	return m, nil
}

type TestMongoBuilder struct {
	query Query
}

func NewTestMongoBuilder(query Query) *TestMongoBuilder {
	return &TestMongoBuilder{query: query}
}

func (t *TestMongoBuilder) Query() Query {
	return t.query
}

func (t *TestMongoBuilder) Build() (interface{}, error) {
	return mongoBuildQuery(t.query)
}

func mongoBuildTermQuery(q Query) (bson.D, error) {
	if q.Type() != TermQueryType {
		return bson.D{}, InvalidQueryType.New()
	}

	tq := q.(TermQuery)

	v, err := toMongoValue(tq.Term().Value())
	if err != nil {
		return bson.D{}, err
	}

	var fq bson.D
	switch tq.Operator() {
	case NOTIN:
		fq = bson.D{{"$not", bson.D{{"$in", v}}}}
	default:
		operator, err := mongoOperator(tq.Operator())
		if err != nil {
			return bson.D{}, err
		}

		fq = bson.D{{operator, v}}
	}

	return bson.D{{tq.Term().Field(), fq}}, nil
}

func mongoBuildConjunctionQuery(q Query) (bson.D, error) {
	if q.Type() != ConjunctionQueryType {
		return bson.D{}, InvalidQueryType.New()
	}

	var dd bson.A
	for _, i := range q.Queries() {
		o, err := mongoBuildQuery(i)
		if err != nil {
			return bson.D{}, err
		}
		dd = append(dd, o)
	}

	tq := q.(ConjunctionQuery)
	conjunction, err := mongoConjunction(tq.Conjunction())
	if err != nil {
		return bson.D{}, err
	}

	return bson.D{{
		conjunction,
		dd,
	}}, nil
}

func mongoBuildQuery(q Query) (bson.D, error) {
	if q.Type() == TermQueryType {
		return mongoBuildTermQuery(q)
	} else if q.Type() == ConjunctionQueryType {
		return mongoBuildConjunctionQuery(q)
	}

	return bson.D{}, InvalidQueryType.New()
}

func toMongoValue(v Value) (interface{}, error) {
	switch v.Hint() {
	case Bool:
	case Int:
	case Int8:
	case Int16:
	case Int32:
	case Int64:
	case Uint:
	case Uint8:
	case Uint16:
	case Uint32:
	case Uint64:
	case Float32, Float64:
	case Complex64, Complex128:
	case String:
	case Time:
	case Duration:
	case Array, Slice:
		var n []interface{}
		t := reflect.ValueOf(v.Value())
		for i := 0; i < t.Len(); i++ {
			x, err := toMongoValue(t.Index(i).Interface().(Value))
			if err != nil {
				return nil, err
			}
			n = append(n, x)
		}

		return n, nil
	default:
		return nil, NotSupportedValueInMongo.New()
	}

	return v.Value(), nil
}
