package graphqlapiv1

import (
	"strconv"

	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/element"
)

var AccountType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Account",
	Fields: graphql.Fields{
		"address": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ac, ok := p.Source.(element.Account)
				if !ok {
					return nil, SourceNotFound.New()
				}

				return ac.Address, nil
			},
		},
		"balance": &graphql.Field{
			Type: AmountType,
			Args: graphql.FieldConfigArgument{
				"unit": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "unit",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ac, ok := p.Source.(element.Account)
				if !ok {
					return nil, SourceNotFound.New()
				}

				if ParseUnitArgument(p, "GON") == "BOS" {
					return common.GonToBOS(ac.Balance), nil
				}

				return ac.Balance, nil
			},
		},
		"sequence_id": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ac, ok := p.Source.(element.Account)
				if !ok {
					return nil, SourceNotFound.New()
				}

				return strconv.FormatUint(ac.SequenceID, 10), nil
			},
		},
		"linked": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ac, ok := p.Source.(element.Account)
				if !ok {
					return nil, SourceNotFound.New()
				}

				return ac.Linked, nil
			},
		},
		"created_height": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ac, ok := p.Source.(element.Account)
				if !ok {
					return nil, SourceNotFound.New()
				}

				return strconv.FormatUint(ac.CreatedHeight, 10), nil
			},
		},
	},
})
