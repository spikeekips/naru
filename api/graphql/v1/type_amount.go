package graphqlapiv1

import (
	"fmt"

	sebakcommon "boscoin.io/sebak/lib/common"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"

	"github.com/spikeekips/naru/common"
)

var AmountType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Amount",
	Description: "Amount type",
	Serialize:   amountToString,
	ParseValue:  stringToAmount,
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			return valueAST.Value
		case *ast.IntValue:
			return valueAST.Value
		}
		return nil
	},
})

func amountToString(a interface{}) interface{} {
	switch a.(type) {
	case string:
		return a
	case common.Amount:
		return a.(common.Amount).String()
	case sebakcommon.Amount:
		return a.(sebakcommon.Amount).String()
	}

	return fmt.Sprintf("%v", a)
}

func stringToAmount(v interface{}) interface{} {
	switch v.(type) {
	case string:
		a, err := sebakcommon.AmountFromString(v.(string))
		if err != nil {
			return nil
		}

		return a
	default:
		return nil
	}
}
