package mongoelement

import (
	"context"

	sebakstorage "boscoin.io/sebak/lib/storage"
	"go.mongodb.org/mongo-driver/bson"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/element"
	"github.com/spikeekips/naru/storage"
	mongostorage "github.com/spikeekips/naru/storage/backend/mongo"
)

type Potion struct {
	s *mongostorage.Storage
}

func NewPotion(s *mongostorage.Storage) Potion {
	return Potion{s: s}
}

func (g Potion) Storage() storage.Storage {
	return g.s
}

func (g Potion) Account(address string) (element.Account, error) {
	var ac element.Account
	err := g.s.Get(element.GetAccountKey(address), &ac)
	return ac, err
}

func (g Potion) Block(hash string) (element.Block, error) {
	var block element.Block
	if err := g.s.Get(element.GetBlockKey(hash), &block); err != nil {
		return element.Block{}, err
	}

	return block, nil
}

func (g Potion) BlockByHeight(height uint64) (element.Block, error) {
	col, err := g.s.Collection(element.BlockPrefix)
	if err != nil {
		return element.Block{}, err
	}

	r := col.FindOne(context.Background(), bson.M{mongostorage.DocField("block.header.height"): height})
	if err := r.Err(); err != nil {
		return element.Block{}, err
	}

	raw, err := r.DecodeBytes()
	if err != nil {
		return element.Block{}, err
	}

	var block element.Block
	_, err = mongostorage.UnmarshalDocument(raw, &block)

	return block, err
}

func (g Potion) LastBlock() (element.Block, error) {
	col, err := g.s.Collection(element.BlockPrefix)
	if err != nil {
		return element.Block{}, err
	}

	cur, err := col.Find(
		context.Background(),
		bson.M{},
		mongooptions.Find().
			SetSort(bson.M{mongostorage.DocField("block.header.height"): -1}).
			SetLimit(1),
	)
	if err != nil {
		return element.Block{}, err
	}
	defer cur.Close(context.Background())

	if !cur.Next(context.Background()) {
		return element.Block{}, err
	}
	if err := cur.Err(); err != nil {
		return element.Block{}, err
	}

	var block element.Block
	_, err = mongostorage.UnmarshalDocument([]byte(cur.Current), &block)
	if err != nil {
		return element.Block{}, err
	}

	return block, err
}

func (g Potion) BlocksIterator(
	iterFunc func() (sebakstorage.IterItem, bool),
	closeFunc func(),
) (
	func() (element.Block, bool, []byte),
	func(),
) {

	return (func() (element.Block, bool, []byte) {
			it, hasNext := iterFunc()
			if !hasNext {
				return element.Block{}, false, []byte{}
			}

			var hash string
			if err := storage.Deserialize(it.Value, &hash); err != nil {
				return element.Block{}, false, []byte{}
			}

			b, err := g.Block(hash)
			if err != nil {
				return element.Block{}, false, []byte{}
			}

			return b, hasNext, it.Key
		}), (func() {
			closeFunc()
		})
}

func (g Potion) Operation(hash string) (op element.Operation, err error) {
	err = g.s.Get(element.GetOperationKey(hash), &op)
	return
}

func (g Potion) OperationsByAccount(address string, options storage.ListOptions) (
	func() (element.Operation, bool, []byte),
	func(),
) {
	nullIterFunc := func() (element.Operation, bool, []byte) {
		return element.Operation{}, false, nil
	}
	nullCloseFunc := func() {}

	col, err := g.s.Collection(element.OperationPrefix)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	q := bson.M{
		"$or": bson.A{
			bson.M{mongostorage.DocField("source"): address},
			bson.M{mongostorage.DocField("target"): address},
		},
	}

	reverse := 1
	if options.Reverse() {
		reverse = -1
	}

	if len(options.Cursor()) > 0 {
		dir := "$lt"
		if options.Reverse() {
			dir = "$gt"
		}

		q = bson.M{
			"$and": bson.A{
				bson.M{mongostorage.DocField("hash"): bson.M{dir: string(options.Cursor())}},
				q,
			}}
	}

	cur, err := col.Find(
		context.Background(),
		q,
		mongooptions.Find().
			SetSort(bson.M{mongostorage.DocField("height"): reverse}).
			SetLimit(int64(options.Limit())),
	)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}
	defer cur.Close(context.Background())

	return func() (element.Operation, bool, []byte) {
			next := cur.Next(context.Background())
			if !next {
				defer cur.Close(context.Background())
				return element.Operation{}, false, nil
			}

			b := []byte(cur.Current)
			var operation element.Operation
			_, err = mongostorage.UnmarshalDocument(b, &operation)
			return operation, true, b
		},
		func() {
			cur.Close(context.Background())
		}
}

func (g Potion) ExistsTransaction(hash string) (bool, error) {
	return g.s.Has(element.GetTransactionKey(hash))
}

func (g Potion) Transaction(hash string) (tx element.Transaction, err error) {
	err = g.s.Get(element.GetTransactionKey(hash), &tx)
	return
}
