package graphqlapiv1

import (
	"errors"
	"strings"

	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
)

func GetPotionFromParams(p graphql.ResolveParams) (element.Potion, error) {
	potion, ok := p.Context.Value("potion").(element.Potion)
	if !ok {
		return nil, PotionIsMissing.New()
	}

	return potion, nil
}

type ListOptonsArgument graphql.FieldConfigArgument

func NewListOptonsArgument() ListOptonsArgument {
	return ListOptonsArgument{
		"reverse": &graphql.ArgumentConfig{
			Type:        graphql.Boolean,
			Description: "sort reverse",
		},
		"cursor": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "iterator starting point",
		},
		"limit": &graphql.ArgumentConfig{
			Type:        graphql.Int,
			Description: "size of result",
		},
	}
}

func (l ListOptonsArgument) Add(name string, config *graphql.ArgumentConfig) ListOptonsArgument {
	l[name] = config

	return l
}

func (l ListOptonsArgument) Done() graphql.FieldConfigArgument {
	return graphql.FieldConfigArgument(l)
}

func (l ListOptonsArgument) ListOptions(p graphql.ResolveParams) (storage.ListOptions, error) {
	var (
		reverse bool
		cursor  string
		limit   int = 50
	)

	if a, ok := p.Args["reverse"]; !ok {
		//
	} else if reverse, ok = a.(bool); !ok {
		return nil, errors.New("invalid `reverse` value found")
	}

	if a, ok := p.Args["cursor"]; !ok {
		//
	} else if cursor, ok = a.(string); !ok {
		return nil, errors.New("invalid `cursor` value found")
	}

	if a, ok := p.Args["limit"]; !ok {
		//
	} else if limit, ok = a.(int); !ok {
		return nil, errors.New("invalid `limit` value found")
	} else if limit < 1 {
		return nil, errors.New("invalid `limit` value found")
	}

	return storage.NewDefaultListOptions(reverse, []byte(cursor), uint64(limit)), nil
}

func ParseUnitArgument(p graphql.ResolveParams, defaultValue string) string {
	var unit string = defaultValue
	if e, ok := p.Args["unit"]; !ok {
		//
	} else if a, ok := e.(string); !ok {
		//
	} else if strings.ToUpper(a) != "GON" && strings.ToUpper(a) != "BOS" {
		//
	} else {
		unit = strings.ToUpper(a)
	}

	return unit
}
