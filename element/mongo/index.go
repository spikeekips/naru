package mongoelement

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/spikeekips/naru/element"
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
			Keys: bson.M{"_v.hash": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_block_hash"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.header.prevblockhash": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_block_header_prevblockhash"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.header.height": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_block_header_height"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.transactions": 1},
			Options: mongooptions.Index().
				SetName("_naru_v0_block_transactions"),
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
			Keys: bson.M{"_v.address": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_account_address"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.balance": -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_account_balance"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.createdheight": -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_account_createdheight"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.linked": -1},
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
			Keys: bson.M{"_v.hash": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_transaction_hash"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.block": 1},
			Options: mongooptions.Index().
				SetName("_naru_v0_transaction_block"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.source": 1, "_v.block": 1},
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
			Keys: bson.M{"_v.hash": 1},
			Options: mongooptions.Index().
				SetUnique(true).
				SetName("_naru_v0_operation_hash"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.txhash": -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_operation_txhash"),
		},
		mongo.IndexModel{
			Keys: bson.M{"_v.txhash": -1, "_v.hash": -1},
			Options: mongooptions.Index().
				SetName("_naru_v0_operation_txhash_hash"),
		},
	},
}
