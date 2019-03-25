package mongoitem

import (
	"context"

	sebakstorage "boscoin.io/sebak/lib/storage"
	"go.mongodb.org/mongo-driver/bson"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/storage"
	mongostorage "github.com/spikeekips/naru/storage/backend/mongo"
	"github.com/spikeekips/naru/storage/item"
)

type Getter struct {
	s *mongostorage.Storage
}

func NewGetter(s *mongostorage.Storage) Getter {
	return Getter{s: s}
}

func (g Getter) Storage() storage.Storage {
	return g.s
}

func (g Getter) Account(address string) (item.Account, error) {
	var ac item.Account
	err := g.s.Get(item.GetAccountKey(address), &ac)
	return ac, err
}

func (g Getter) Block(hash string) (item.Block, error) {
	var block item.Block
	if err := g.s.Get(item.GetBlockKey(hash), &block); err != nil {
		return item.Block{}, err
	}

	return block, nil
}

func (g Getter) BlockByHeight(height uint64) (item.Block, error) {
	col, err := g.s.Collection(item.BlockPrefix)
	if err != nil {
		return item.Block{}, err
	}

	r := col.FindOne(context.Background(), bson.M{mongostorage.DocField("block.header.height"): height})
	if err := r.Err(); err != nil {
		return item.Block{}, err
	}

	raw, err := r.DecodeBytes()
	if err != nil {
		return item.Block{}, err
	}

	var block item.Block
	_, err = mongostorage.UnmarshalDocument(raw, &block)

	return block, err
}

func (g Getter) LastBlock() (item.Block, error) {
	col, err := g.s.Collection(item.BlockPrefix)
	if err != nil {
		return item.Block{}, err
	}

	cur, err := col.Find(
		context.Background(),
		bson.M{},
		mongooptions.Find().
			SetSort(bson.M{mongostorage.DocField("block.header.height"): -1}).
			SetLimit(1),
	)
	if err != nil {
		return item.Block{}, err
	}
	defer cur.Close(context.Background())

	if !cur.Next(context.Background()) {
		return item.Block{}, err
	}
	if err := cur.Err(); err != nil {
		return item.Block{}, err
	}

	var block item.Block
	_, err = mongostorage.UnmarshalDocument([]byte(cur.Current), &block)
	if err != nil {
		return item.Block{}, err
	}

	return block, err
}

func (g Getter) BlocksIterator(
	iterFunc func() (sebakstorage.IterItem, bool),
	closeFunc func(),
) (
	func() (item.Block, bool, []byte),
	func(),
) {

	return (func() (item.Block, bool, []byte) {
			it, hasNext := iterFunc()
			if !hasNext {
				return item.Block{}, false, []byte{}
			}

			var hash string
			if err := storage.Deserialize(it.Value, &hash); err != nil {
				return item.Block{}, false, []byte{}
			}

			b, err := g.Block(hash)
			if err != nil {
				return item.Block{}, false, []byte{}
			}

			return b, hasNext, it.Key
		}), (func() {
			closeFunc()
		})
}

func (g Getter) Operation(hash string) (op item.Operation, err error) {
	err = g.s.Get(item.GetOperationKey(hash), &op)
	return
}

func (g Getter) OperationsByAccount(address string, options storage.ListOptions) (
	func() (item.Operation, bool, []byte),
	func(),
) {
	nullIterFunc := func() (item.Operation, bool, []byte) {
		return item.Operation{}, false, nil
	}
	nullCloseFunc := func() {}

	col, err := g.s.Collection(item.OperationPrefix)
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

	return func() (item.Operation, bool, []byte) {
			next := cur.Next(context.Background())
			if !next {
				defer cur.Close(context.Background())
				return item.Operation{}, false, nil
			}

			b := []byte(cur.Current)
			var operation item.Operation
			_, err = mongostorage.UnmarshalDocument(b, &operation)
			return operation, true, b
		},
		func() {
			cur.Close(context.Background())
		}
}

func (g Getter) ExistsTransaction(hash string) (bool, error) {
	return g.s.Has(item.GetTransactionKey(hash))
}

func (g Getter) Transaction(hash string) (tx item.Transaction, err error) {
	err = g.s.Get(item.GetTransactionKey(hash), &tx)
	return
}
