package graphqlapiv1

import (
	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/element"
)

func getOperationBySource(p graphql.ResolveParams) (element.Operation, error) {
	operation, ok := p.Source.(element.Operation)
	if !ok {
		return element.Operation{}, SourceNotFound.New()
	}

	return operation, nil
}

var OperationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Operation",
	Fields: graphql.Fields{
		"hash": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.Hash, nil
			},
		},
		"op_hash": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.OpHash, nil
			},
		},
		"op_index": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.OpIndex, nil
			},
		},
		"tx_index": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.TxHash, nil
			},
		},
		"type": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.Type, nil
			},
		},
		"source": &graphql.Field{
			Type: AccountType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				potion, err := GetPotionFromParams(p)
				if err != nil {
					return nil, err
				}

				return potion.Account(operation.Source)
			},
		},
		"target": &graphql.Field{
			Type: AccountType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				potion, err := GetPotionFromParams(p)
				if err != nil {
					return nil, err
				}

				return potion.Account(operation.Target)
			},
		},
		"block": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.Block, nil
			},
		},
		"sequence_id": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.SequenceID, nil
			},
		},
		"linked": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.Linked, nil
			},
		},
		"amount": &graphql.Field{
			Type: AmountType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return operation.Amount, nil
			},
		},
		"raw": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				operation, err := getOperationBySource(p)
				if err != nil {
					return nil, err
				}

				return string(operation.Raw), nil
			},
		},
	},
})
