package graphqlapiv1

import (
	sebakcommon "boscoin.io/sebak/lib/common"
	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
)

func getTransactionBySource(p graphql.ResolveParams) (element.Transaction, error) {
	transaction, ok := p.Source.(element.Transaction)
	if !ok {
		return element.Transaction{}, SourceNotFound.New()
	}

	return transaction, nil
}

var TransactionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Transaction",
	Fields: graphql.Fields{
		"hash": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return transaction.Hash, nil
			},
		},
		"block": &graphql.Field{
			Type: BlockType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				potion, err := GetPotionFromParams(p)
				if err != nil {
					return nil, err
				}
				block, err := potion.Block(transaction.Block)
				if err != nil {
					return nil, err
				}

				return block, nil
			},
		},
		"sequence_id": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return transaction.SequenceID, nil
			},
		},
		"signature": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return transaction.Signature, nil
			},
		},
		"source": &graphql.Field{
			Type: AccountType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				potion, err := GetPotionFromParams(p)
				if err != nil {
					return nil, err
				}

				return potion.Account(transaction.Source)
			},
		},
		"fee": &graphql.Field{
			Type: AmountType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return transaction.Fee, nil
			},
		},
		"operations": &graphql.Field{
			Type: graphql.NewList(OperationType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				potion, err := GetPotionFromParams(p)
				if err != nil {
					return nil, err
				}

				iterFunc, closeFunc := potion.OperationsByTransaction(
					transaction.Hash,
					storage.NewDefaultListOptions(false, nil, 0),
				)

				defer closeFunc()

				var operations []element.Operation
				for {
					operation, next, _ := iterFunc()
					if !next {
						break
					}
					operations = append(operations, operation)
				}

				return operations, nil
			},
		},
		"amount": &graphql.Field{
			Type: AmountType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return transaction.Amount, nil
			},
		},
		"confirmed": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return sebakcommon.ParseISO8601(transaction.Confirmed)
			},
		},
		"created": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return sebakcommon.ParseISO8601(transaction.Created)
			},
		},
		"raw": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				transaction, err := getTransactionBySource(p)
				if err != nil {
					return nil, err
				}

				return string(transaction.Raw), nil
			},
		},
	},
})
