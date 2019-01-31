package storage

const (
	InternalPrefix            = "\x00\x00" // internal data
	BlockPrefix               = "\x10\x00" // block
	BlockHeightPrefix         = "\x10\x01"
	TransactionPrefix         = "\x20\x00" // transaction
	TransactionSourcePrefix   = "\x20\x01"
	TransactionAccountsPrefix = "\x20\x02"
	TransactionBlockPrefix    = "\x20\x03"
	AccountPrefix             = "\x30\x00" // account
)
