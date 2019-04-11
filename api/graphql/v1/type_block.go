package graphqlapiv1

import (
	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
)

func getBlockBySource(p graphql.ResolveParams) (element.Block, error) {
	block, ok := p.Source.(element.Block)
	if !ok {
		return element.Block{}, SourceNotFound.New()
	}

	return block, nil
}

var BlockType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Block",
	Fields: graphql.Fields{
		"version": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Header.Version, nil
			},
		},
		"prev_block_hash": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Header.PrevBlockHash, nil
			},
		},
		"transactions_root": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Header.TransactionsRoot, nil
			},
		},
		"proposed_time": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Header.ProposedTime, nil
			},
		},
		"height": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Header.Height, nil
			},
		},
		"total_txs": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Header.TotalTxs, nil
			},
		},
		"total_ops": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Header.TotalOps, nil
			},
		},
		"transaction_hashes": &graphql.Field{
			// TODO should be TransactionType
			Type: graphql.NewList(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Transactions, nil
			},
		},
		"transactions": &graphql.Field{
			// TODO should be TransactionType
			Type: graphql.NewList(TransactionType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				potion, err := GetPotionFromParams(p)
				if err != nil {
					return nil, err
				}

				var transactions []element.Transaction
				iterFunc, closeFunc := potion.TransactionsByBlock(block.Hash, storage.NewDefaultListOptions(false, nil, 0))
				defer closeFunc()

				for {
					transaction, next, _ := iterFunc()
					if !next {
						break
					}
					transactions = append(transactions, transaction)
				}

				return transactions, nil
			},
		},
		"proposer_transaction_hash": &graphql.Field{
			// TODO should be TransactionType
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.ProposerTransaction, nil
			},
		},
		"proposer_transaction": &graphql.Field{
			Type: TransactionType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				potion, err := GetPotionFromParams(p)
				if err != nil {
					return nil, err
				}
				return potion.Transaction(block.ProposerTransaction)
			},
		},
		"hash": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Hash, nil
			},
		},
		"proposer": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Proposer, nil
			},
		},
		"round": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Round, nil
			},
		},
		"confirmed": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				block, err := getBlockBySource(p)
				if err != nil {
					return nil, err
				}

				return block.Confirmed, nil
			},
		},
	},
})
