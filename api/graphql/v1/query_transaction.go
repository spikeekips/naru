package graphqlapiv1

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/spikeekips/naru/element"
)

var GetTransactionQuery *graphql.Field = &graphql.Field{
	Type: TransactionType,
	Args: graphql.FieldConfigArgument{
		"hash": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "transaction hash",
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		var hash string
		if e, ok := p.Args["hash"]; !ok {
			return nil, errors.New("`hash` argument is missing")
		} else if h, ok := e.(string); !ok {
			return nil, errors.New("invalid `hash` value found")
		} else {
			hash = h
		}

		potion, err := GetPotionFromParams(p)
		if err != nil {
			return nil, err
		}

		transaction, err := potion.Transaction(hash)
		if err != nil {
			return nil, err
		}

		return transaction, nil
	},
}

var GetTransactionsQuery *graphql.Field = &graphql.Field{
	Type: graphql.NewList(TransactionType),
	Args: NewListOptonsArgument().
		Add(
			"block",
			&graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "block of transaction",
			},
		).
		Add(
			"account",
			&graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "transactions of account",
			},
		).
		Done(),
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		var blockHash string
		if e, ok := p.Args["block"]; !ok {
			//
		} else if h, ok := e.(string); !ok {
			return nil, errors.New("invalid `hash` value found")
		} else {
			blockHash = h
		}

		var account string
		if e, ok := p.Args["account"]; !ok {
			//
		} else if h, ok := e.(string); !ok {
			return nil, errors.New("invalid `account value found")
		} else {
			account = h
		}

		potion, err := GetPotionFromParams(p)
		if err != nil {
			return nil, err
		}

		options, err := ListOptonsArgument{}.ListOptions(p)
		if err != nil {
			return nil, err
		}

		var iterFunc func() (element.Transaction, bool, []byte)
		var closeFunc func()
		if len(blockHash) > 0 {
			if _, err = potion.Block(blockHash); err != nil {
				return nil, err
			}

			iterFunc, closeFunc = potion.TransactionsByBlock(blockHash, options)
		} else if len(account) > 0 {
			iterFunc, closeFunc = potion.TransactionsByAccount(account, options)
		}
		defer closeFunc()

		var transactions []element.Transaction
		for {
			tx, next, _ := iterFunc()
			if !next {
				break
			}
			transactions = append(transactions, tx)
		}

		return transactions, nil
	},
}
