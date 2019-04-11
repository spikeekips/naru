package graphqlapiv1

import (
	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/element"
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
		"transactions": &graphql.Field{
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
		"proposer_transaction": &graphql.Field{
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
