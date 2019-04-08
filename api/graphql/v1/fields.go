package graphqlapiv1

/*

import (
	"strconv"
	"strings"

	"boscoin.io/sebak/lib/common/keypair"
	"github.com/graphql-go/graphql"

	"github.com/spikeekips/naru/element"
)

var TotalSupply *graphql.Field = &graphql.Field{
	Type: graphql.String,
	Args: graphql.FieldConfigArgument{
		"excludes": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "This empty space seperated addresses will be excluded from total supply",
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		var potion element.Potion
		if source, ok := p.Source.(map[string]interface{}); !ok {
			return nil, PotionIsMissing.New()
		} else if po, ok := source["potion"]; !ok {
			return nil, PotionIsMissing.New()
		} else if potion, ok = po.(element.Potion); !ok {
			return nil, PotionIsMissing.New()
		}

		bs, err := potion.BlockStat()
		if err != nil {
			return nil, err
		}

		var excludes []string
		if e, ok := p.Args["excludes"]; !ok {
			//
		} else if es, ok := e.(string); !ok {
			//
		} else {
			excludes = strings.Fields(es)
			for _, a := range excludes {
				if _, err = keypair.Parse(a); err != nil {
					return nil, InValidPublicAddress.New()
				}
			}
		}

		if len(excludes) < 1 {
			return strconv.FormatUint(bs.TotalSupply, 10), nil
		}

		var amount uint64
		for _, address := range excludes {
			ac, err := potion.Account(address)
			if err != nil {
				log.Error("account not found", "error", err)
				continue
			}

			amount += uint64(ac.Balance)
		}

		return strconv.FormatUint(bs.TotalSupply-amount, 10), nil
	},
}
*/
