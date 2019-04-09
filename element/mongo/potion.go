package mongoelement

import (
	"context"
	"strings"

	logging "github.com/inconshreveable/log15"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (g Potion) checkOne(prefix string, dryRun bool) error {
	collectionName, err := mongostorage.GetCollection(prefix)
	if err != nil {
		return err
	}

	log_ := log.New(logging.Ctx{"prefix": collectionName, "dryRun": dryRun})

	col, err := g.s.Collection(prefix)
	if err != nil {
		return err
	}
	indexes := col.Indexes()
	cur, err := indexes.List(context.Background())
	if err != nil {
		return err
	}

	var foundNames []string
	for {
		if !cur.Next(context.Background()) {
			break
		}
		var i bson.M
		if err := cur.Decode(&i); err != nil {
			return err
		}
		n := i["name"].(string)
		if !strings.HasPrefix(n, "_naru_") {
			continue
		}

		foundNames = append(foundNames, n)
	}
	cur.Close(context.Background())
	log_.Debug("these indices found", "indexes", foundNames)

	var willAdd, willRemove []string
	for _, i := range allIndexes[prefix] {
		var found bool
		for _, j := range foundNames {
			if j == *i.Options.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}
		willAdd = append(willAdd, *i.Options.Name)
	}

	for _, j := range foundNames {
		var found bool
		for _, i := range allIndexes[prefix] {
			if j == *i.Options.Name {
				found = true
				break
			}
		}
		if found {
			continue
		}
		willRemove = append(willRemove, j)
	}

	log_.Debug("these indices will be removed", "indexes", willRemove)

	if !dryRun && len(willRemove) > 0 {
		for _, r := range willRemove {
			if _, err := indexes.DropOne(context.Background(), r); err != nil {
				return err
			}
		}
	}

	log_.Debug("these indices will be added", "indexes", willAdd)
	var addIndexes []mongo.IndexModel
	for _, a := range willAdd {
		for _, i := range allIndexes[prefix] {
			if a != *i.Options.Name {
				continue
			}
			addIndexes = append(addIndexes, i)
		}
	}

	if !dryRun && len(addIndexes) > 0 {
		if _, err := indexes.CreateMany(context.Background(), addIndexes); err != nil {
			return err
		}
	}

	return nil
}

func (g Potion) Check() error {
	prefixes := []string{
		element.InternalPrefix,
		element.BlockPrefix,
		element.TransactionPrefix,
		element.AccountPrefix,
		element.OperationPrefix,
	}

	for _, prefix := range prefixes {
		if err := g.checkOne(prefix, false); err != nil {
			return err
		}
	}

	return nil
}

func (g Potion) Storage() storage.Storage {
	return g.s
}

func (g Potion) Account(address string) (element.Account, error) {
	var ac element.Account
	err := g.s.Get(element.GetAccountKey(address), &ac)
	return ac, err
}

func (g Potion) Accounts(sort string, options storage.ListOptions) (
	func() (element.Account, bool, []byte),
	func(),
) {
	nullIterFunc := func() (element.Account, bool, []byte) {
		return element.Account{}, false, nil
	}
	nullCloseFunc := func() {}

	col, err := g.s.Collection(element.AccountPrefix)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	q := bson.M{}

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
				bson.M{mongostorage.DocField("address"): bson.M{dir: string(options.Cursor())}},
				q,
			}}
	}

	if len(sort) < 1 {
		sort = mongostorage.DocField("createdheight")
	}

	cur, err := col.Find(
		context.Background(),
		q,
		mongooptions.Find().
			SetSort(bson.D{{mongostorage.DocField(sort), reverse}, {"_id", reverse}}).
			SetLimit(int64(options.Limit())),
	)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	return func() (element.Account, bool, []byte) {
			next := cur.Next(context.Background())
			if !next {
				defer cur.Close(context.Background())
				return element.Account{}, false, nil
			}

			b := []byte(cur.Current)
			var account element.Account
			_, err = mongostorage.UnmarshalDocument(b, &account)
			return account, true, b
		},
		func() {
			cur.Close(context.Background())
		}
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

func (g Potion) BlocksByHeight(start, end uint64) (
	func() (element.Block, bool, []byte),
	func(),
) {
	nullIterFunc := func() (element.Block, bool, []byte) {
		return element.Block{}, false, nil
	}
	nullCloseFunc := func() {}

	col, err := g.s.Collection(element.BlockPrefix)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	q := bson.M{
		"$and": bson.A{
			bson.M{mongostorage.DocField("block.header.height"): bson.M{"$gte": start}},
			bson.M{mongostorage.DocField("block.header.height"): bson.M{"$lt": end}},
		},
	}

	cur, err := col.Find(
		context.Background(),
		q,
		mongooptions.Find().
			SetSort(bson.M{mongostorage.DocField("block.header.height"): 1}),
	)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	return func() (element.Block, bool, []byte) {
			next := cur.Next(context.Background())
			if !next {
				defer cur.Close(context.Background())
				return element.Block{}, false, nil
			}

			b := []byte(cur.Current)
			var block element.Block
			_, err = mongostorage.UnmarshalDocument(b, &block)
			return block, true, b
		},
		func() {
			cur.Close(context.Background())
		}
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

func (g Potion) OperationsByTransaction(hash string, options storage.ListOptions) (
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

	q := bson.M{mongostorage.DocField("txhash"): hash}

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

	reverse := 1
	if options.Reverse() {
		reverse = -1
	}

	cur, err := col.Find(
		context.Background(),
		q,
		mongooptions.Find().
			SetSort(bson.M{mongostorage.DocField("hash"): reverse}).
			SetLimit(int64(options.Limit())),
	)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

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

func (g Potion) TransactionsByBlock(hash string, options storage.ListOptions) (
	func() (element.Transaction, bool, []byte),
	func(),
) {
	nullIterFunc := func() (element.Transaction, bool, []byte) {
		return element.Transaction{}, false, nil
	}
	nullCloseFunc := func() {}

	block, err := g.Block(hash)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	col, err := g.s.Collection(element.TransactionPrefix)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	q := bson.M{
		mongostorage.DocField("hash"): bson.M{
			"$in": block.Transactions,
		},
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

	reverse := 1
	if options.Reverse() {
		reverse = -1
	}

	cur, err := col.Find(
		context.Background(),
		q,
		mongooptions.Find().
			SetSort(bson.M{mongostorage.DocField("hash"): reverse}).
			SetLimit(int64(options.Limit())),
	)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	return func() (element.Transaction, bool, []byte) {
			next := cur.Next(context.Background())
			if !next {
				defer cur.Close(context.Background())
				return element.Transaction{}, false, nil
			}

			b := []byte(cur.Current)
			var transaction element.Transaction
			_, err = mongostorage.UnmarshalDocument(b, &transaction)
			return transaction, true, b
		},
		func() {
			cur.Close(context.Background())
		}
}

func (g Potion) OperationsByHeight(start, end uint64) (
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
		"$and": bson.A{
			bson.M{mongostorage.DocField("height"): bson.M{"$gte": start}},
			bson.M{mongostorage.DocField("height"): bson.M{"$lt": end}},
		},
	}

	cur, err := col.Find(
		context.Background(),
		q,
		mongooptions.Find().
			SetSort(bson.M{mongostorage.DocField("height"): 1}),
	)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

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

func (g Potion) BlockStat() (element.BlockStat, error) {
	var bs element.BlockStat
	err := g.s.Get(element.GetBlockStatKey(), &bs)
	return bs, err
}

func (g Potion) TransactionsByAccount(address string, options storage.ListOptions) (
	func() (element.Transaction, bool, []byte),
	func(),
) {
	nullIterFunc := func() (element.Transaction, bool, []byte) {
		return element.Transaction{}, false, nil
	}
	nullCloseFunc := func() {}

	col, err := g.s.Collection(element.TransactionPrefix)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	q := bson.M{mongostorage.DocField("source"): address}

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
			SetSort(bson.M{mongostorage.DocField("confirmed"): reverse}).
			SetLimit(int64(options.Limit())),
	)
	if err != nil {
		return nullIterFunc, nullCloseFunc
	}

	return func() (element.Transaction, bool, []byte) {
			next := cur.Next(context.Background())
			if !next {
				defer cur.Close(context.Background())
				return element.Transaction{}, false, nil
			}

			b := []byte(cur.Current)
			var transaction element.Transaction
			_, err = mongostorage.UnmarshalDocument(b, &transaction)
			return transaction, true, b
		},
		func() {
			cur.Close(context.Background())
		}
}
