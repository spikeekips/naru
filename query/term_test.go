package query

import (
	"testing"
	"time"

	"boscoin.io/sebak/lib/common"
	"github.com/stretchr/testify/suite"
)

type testTerm struct {
	suite.Suite
}

func (t *testTerm) TestValueHint() {
	field := "this-is-field"
	value := "string-value"
	tm, err := NewTerm(field, value)

	t.NoError(err)
	t.Equal(field, tm.Field())
	t.Equal(value, tm.Value().Value())
	t.Equal(String, tm.Value().Hint())
}

func (t *testTerm) TestInvalidValueHint() {
	field := "this-is-field"

	var value *int
	_, err := NewTerm(field, value)

	t.True(InvaludValue.Equal(err))
}

func (t *testTerm) TestValueHintDuration() {
	field := "this-is-field"
	value := time.Second * 1
	tm, err := NewTerm(field, value)

	t.NoError(err)
	t.Equal(field, tm.Field())
	t.Equal(value, tm.Value().Value())
	t.Equal(Duration, tm.Value().Hint())
}

func (t *testTerm) TestValueHintTime() {
	field := "this-is-time-field"
	value := time.Now()
	tm, err := NewTerm(field, value)

	t.NoError(err)
	t.Equal(field, tm.Field())
	t.Equal(value, tm.Value().Value())
	t.Equal(Time, tm.Value().Hint())
}

func (t *testTerm) TestTimeValueString() {
	field := "this-is-field"
	value, _ := common.ParseISO8601("2019-03-07T16:40:27.513727000+09:00")
	tm, err := NewTerm(field, value)

	t.NoError(err)
	t.Equal(`{"field":"this-is-field","value":{"hint":"time","value":"2019-03-07T16:40:27.513727+09:00"}}`, tm.String())
}

func (t *testTerm) TestDurationValueString() {
	field := "this-is-field"
	value := time.Second * 9
	tm, err := NewTerm(field, value)

	t.NoError(err)
	t.Equal(`{"field":"this-is-field","value":{"hint":"duration","value":9000000000}}`, tm.String())
}

func (t *testTerm) TestArray() {
	field := "this-is-field"
	value := [2]int{33, 44}
	tm, err := NewTerm(field, value)

	t.NoError(err)
	t.Equal(`{"field":"this-is-field","value":{"hint":"array","value":[{"hint":"int","value":33},{"hint":"int","value":44}]}}`, tm.String())
}

func (t *testTerm) TestSlice() {
	field := "this-is-field"
	value := []int{33, 44}
	tm, err := NewTerm(field, value)

	t.NoError(err)
	t.Equal(`{"field":"this-is-field","value":{"hint":"slice","value":[{"hint":"int","value":33},{"hint":"int","value":44}]}}`, tm.String())
}

func TestTerm(t *testing.T) {
	suite.Run(t, new(testTerm))
}
