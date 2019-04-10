package mongoelement

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/element"
	mongostorage "github.com/spikeekips/naru/storage/backend/mongo"
)

var allIndexes = map[string][]mongo.IndexModel{
	element.BlockPrefix: []mongo.IndexModel{
		mongo.IndexModel{
			Keys: bson.M{"_k": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_block_k"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("hash"): 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_block_hash"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("header.prevblockhash"): 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_block_header.prevblockhash"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("header.height"): 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_block_header.height"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("transactions"): 1},
			Options: mongooptions.Index().
				SetName("_naru_v0_block_transactions"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("confirmed"): 1},
			Options: mongooptions.Index().
				SetName("_naru_v0_block_confirmed"),
		},
	},
	element.AccountPrefix: []mongo.IndexModel{
		mongo.IndexModel{
			Keys: bson.M{"_k": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_account_k"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("address"): 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_account_address"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("balance"): -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_account_balance"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("createdheight"): -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_account_createdheight"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("linked"): -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_account_linked"),
		},
	},
	element.TransactionPrefix: []mongo.IndexModel{
		mongo.IndexModel{
			Keys: bson.M{"_k": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_transaction_k"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("hash"): 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_transaction_hash"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("block"): 1},
			Options: mongooptions.Index().
				SetName("_naru_v0_transaction_block"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("source"): 1, mongostorage.DocField("confirmed"): 1},
			Options: mongooptions.Index().
				SetName("_naru_v0_transaction_source"),
		},
	},
	element.OperationPrefix: []mongo.IndexModel{
		mongo.IndexModel{
			Keys: bson.M{"_k": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_operation_k"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("hash"): 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_operation_hash"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("txhash"): -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_operation_txhash"),
		},
		mongo.IndexModel{
			Keys: bson.M{mongostorage.DocField("txhash"): -1, mongostorage.DocField("hash"): -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_operation_txhash_hash"),
		},
	},
}
