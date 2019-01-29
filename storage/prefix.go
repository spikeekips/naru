package storage

const (
	InternalPrefix    = "\x00\x00" // internal data
	BlockPrefix       = "\x10\x00"
	BlockHeightPrefix = "\x10\x01"
	TransactionPrefix = "\x20\x00"
	AccountPrefix     = "\x30\x00"
)
