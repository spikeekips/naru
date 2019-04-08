package graphqlapiv1

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/spikeekips/naru/element"
)

var GetBlockQuery *graphql.Field = &graphql.Field{
	Type: BlockType,
	Args: graphql.FieldConfigArgument{
		"hash": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "block hash",
		},
		"height": &graphql.ArgumentConfig{
			Type:        graphql.Int,
			Description: "block height",
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		potion, err := GetPotionFromParams(p)
		if err != nil {
			return nil, err
		}

		var height uint64
		if e, ok := p.Args["height"]; !ok {
			//return nil, errors.New("`height` argument is missing")
		} else if h, ok := e.(int); !ok {
			return nil, errors.New("invalid `height` value found")
		} else if uh := uint64(h); uh < 0 {
			return nil, InValidArgument.New()
		} else {
			height = uh
		}

		var hash string
		if e, ok := p.Args["hash"]; !ok {
			//return nil, errors.New("`hash` argument is missing")
		} else if h, ok := e.(string); !ok {
			return nil, errors.New("invalid `hash` value found")
		} else {
			hash = h
		}

		var block element.Block
		if len(hash) > 0 {
			block, err = potion.Block(hash)
		} else if height > 0 {
			block, err = potion.BlockByHeight(height)
		} else {
			return nil, errors.New("empty arguments")
		}

		if err != nil {
			return nil, err
		}

		return block, nil
	},
}
