package graphqlapiv1

import (
	"github.com/graphql-go/handler"

	"github.com/spikeekips/naru/element"
)

func Handler(potion element.Potion) *handler.Handler {
	schema := NewSchema()

	return handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
}
