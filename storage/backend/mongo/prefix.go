package mongostorage

import "github.com/spikeekips/naru/storage/item"

var (
	CollectionByPrefix = map[string]string{
		item.InternalPrefix[:2]:    "internal",
		item.BlockPrefix[:2]:       "block",
		item.TransactionPrefix[:2]: "transaction",
		item.AccountPrefix[:2]:     "account",
		item.OperationPrefix[:2]:   "operation",
	}
)

func getCollection(key string) (string, error) {
	c, ok := CollectionByPrefix[key[:2]]
	if !ok {
		return "", UnknownCollection.New()
	}

	return c, nil
}
