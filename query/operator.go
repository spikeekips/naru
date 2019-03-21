package query

import (
	"github.com/spikeekips/naru/common"
)

type Conjunction int

const (
	EMPTYConjunction Conjunction = iota
	AND
	OR
)

func (c Conjunction) String() string {
	switch c {
	case AND:
		return "$AND"
	case OR:
		return "$OR"
	}

	return "$EMPTYCOND"
}

func (c Conjunction) MarshalJSON() ([]byte, error) {
	return common.MarshalJSONNotEscapeHTML(c.String())
}

type Operator int

const (
	EMPTYOperator Operator = iota + 3
	IS
	NOT
	GT
	GTE
	LT
	LTE
	IN
	NOTIN
)

func (o Operator) String() string {
	switch o {
	case IS:
		return "$IS"
	case NOT:
		return "$NOT"
	case GT:
		return "$GT"
	case GTE:
		return "$GTE"
	case LT:
		return "$LT"
	case LTE:
		return "$LTE"
	case IN:
		return "$IN"
	case NOTIN:
		return "$NOTIN"
	}

	return "$EMPTYOP"
}

func (o Operator) MarshalJSON() ([]byte, error) {
	return common.MarshalJSONNotEscapeHTML(o.String())
}
