package graphqlapiv1

import (
	"github.com/graphql-go/graphql"
)

func NewSchema() graphql.Schema {
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
		"total_supply": TotalSupplyQuery,
		"account":      GetAccountQuery,
		"accounts":     GetAccountsQuery,
		"block":        GetBlockQuery,
		"transaction":  GetTransactionQuery,
		"transactions": GetTransactionsQuery,
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Crit("failed to create new schema", "error", err)
		panic(err)
	}

	return schema
}
