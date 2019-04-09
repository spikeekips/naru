package mongostorage

import "github.com/spikeekips/naru/element"

var (
	CollectionByPrefix = map[string]string{
		element.InternalPrefix[:2]:    "internal",
		element.BlockPrefix[:2]:       "block",
		element.TransactionPrefix[:2]: "transaction",
		element.AccountPrefix[:2]:     "account",
		element.OperationPrefix[:2]:   "operation",
	}
)

func GetCollection(key string) (string, error) {
	c, ok := CollectionByPrefix[key[:2]]
	if !ok {
		return "", UnknownCollection.New()
	}

	return c, nil
}
