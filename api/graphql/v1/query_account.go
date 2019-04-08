package graphqlapiv1

import (
	"errors"

	"boscoin.io/sebak/lib/common/keypair"
	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/element"
)

var GetAccountQuery *graphql.Field = &graphql.Field{
	Type: AccountType,
	Args: graphql.FieldConfigArgument{
		"address": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "account address",
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		potion, err := GetPotionFromParams(p)
		if err != nil {
			return nil, err
		}

		var address string
		if e, ok := p.Args["address"]; !ok {
			return nil, errors.New("`address` argument is missing")
		} else if address, ok = e.(string); !ok {
			return nil, errors.New("invalid `address` value found")
		} else if _, err := keypair.Parse(address); err != nil {
			return nil, InValidPublicAddress.New()
		}

		ac, err := potion.Account(address)
		if err != nil {
			return nil, err
		}

		return ac, nil
	},
}

var GetAccountsQuery *graphql.Field = &graphql.Field{
	Type: graphql.NewList(AccountType),
	Args: NewListOptonsArgument().
		Add(
			"sort",
			&graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "sort field",
			},
		).
		Add(
			"balance",
			&graphql.ArgumentConfig{
				Type:        AmountType,
				Description: "balance",
			},
		).Done(),
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		potion, err := GetPotionFromParams(p)
		if err != nil {
			return nil, err
		}

		options, err := ListOptonsArgument{}.ListOptions(p)
		if err != nil {
			return nil, err
		}

		var sort string
		if s, ok := p.Args["sort"]; !ok {
			//
		} else if sort, ok = s.(string); !ok {
			return nil, errors.New("invalid `sort` value found")
		}

		iterFunc, closeFunc := potion.Accounts(sort, options)
		defer closeFunc()

		var accounts []element.Account
		for {
			ac, next, _ := iterFunc()
			if !next {
				break
			}
			accounts = append(accounts, ac)
		}

		return accounts, nil
	},
}
