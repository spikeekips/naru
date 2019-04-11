package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	sebakblock "boscoin.io/sebak/lib/block"
	sebakcommon "boscoin.io/sebak/lib/common"
	sebakrunner "boscoin.io/sebak/lib/node/runner"
	sebakstorage "boscoin.io/sebak/lib/storage"
)

var (
	storageConfig      *sebakstorage.Config
	st                 *sebakstorage.LevelDBBackend
	blockHeight        uint64
	blocksLimit        uint64 = sebakrunner.MaxLimitListOptions
	maxNumberOfWorkers int    = 100
)

func printFlagsError(s string, err error) {
	var errString string
	if err != nil {
		errString = err.Error()
	}

	if len(s) > 0 {
		fmt.Println("error:", s, "", errString)
	}
	fmt.Fprintf(os.Stderr, "Usage: %s <sebak storage> <block height>\n", os.Args[0])

	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.Parse(os.Args[1:])

	if flags.NArg() < 2 {
		printFlagsError("missing arguments", nil)
	}

	{
		var err error
		var storageConfig *sebakstorage.Config
		if storageConfig, err = sebakstorage.NewConfigFromString(flags.Arg(0)); err != nil {
			printFlagsError("<storage>", err)
		}
		if st, err = sebakstorage.NewStorage(storageConfig); err != nil {
			printFlagsError("failed to initialize storage", err)
		}
	}

	{
		var err error
		blockHeight, err = strconv.ParseUint(flags.Arg(1), 10, 64)
		if err != nil {
			printFlagsError("<block height>", err)
		}
	}
}

func removeBlock(hash string, chanError chan<- error) (err error) {
	defer func() {
		chanError <- err
	}()

	var block sebakblock.Block
	if block, err = sebakblock.GetBlock(st, hash); err != nil {
		fmt.Println("error: failed to get block", err)
		return
	}

	// remove block.Block
	if err = st.Remove(fmt.Sprintf("%s%s", sebakcommon.BlockPrefixHash, hash)); err != nil {
		fmt.Println("error: failed to remove BlockPrefixHash", err, block)
		return
	}

	confirmedPrefix := fmt.Sprintf(
		"%s%s-%s",
		sebakcommon.BlockPrefixConfirmed,
		block.ProposedTime,
		sebakcommon.EncodeUint64ToByteSlice(block.Height),
	)
	iterFunc, closeFunc := st.GetIterator(
		confirmedPrefix,
		sebakstorage.NewDefaultListOptions(false, nil, 0),
	)
	for {
		item, next := iterFunc()
		if !next {
			break
		}
		if err = st.Remove(string(item.Key)); err != nil {
			fmt.Println("error: failed to remove confirmed", err, block)
			return
		}
	}
	closeFunc()

	if err = st.Remove(sebakblock.GetBlockKeyPrefixHeight(block.Height)); err != nil {
		fmt.Println("error: failed to remove height", err, block)
		return
	}

	return
}

func removeBlockWorker(wid int, chanWorker <-chan string, chanError chan<- error) {
	for hash := range chanWorker {
		if err := removeBlock(hash, chanError); err != nil {
			break
		}
	}
}

func main() {
	// get latest block
	latestBlock := sebakblock.GetLatestBlock(st)

	fmt.Println("latestBlock", latestBlock)
	if latestBlock.Height == blockHeight {
		fmt.Println("> nothing happened")
		return
	}

	numberOfWorkers := int((latestBlock.Height - blockHeight) / blocksLimit)
	if numberOfWorkers < 1 {
		numberOfWorkers = 1
	} else if numberOfWorkers > maxNumberOfWorkers {
		numberOfWorkers = maxNumberOfWorkers
	}

	chanWorker := make(chan string)
	chanError := make(chan error)

	for wid := 0; wid < numberOfWorkers; wid++ {
		go removeBlockWorker(wid, chanWorker, chanError)
	}

	go func() {
		iterFunc, closeFunc := st.GetIterator(
			sebakcommon.BlockPrefixHeight,
			sebakstorage.NewDefaultListOptions(
				false,
				[]byte(sebakblock.GetBlockKeyPrefixHeight(blockHeight)),
				0,
			),
		)

		var item sebakstorage.IterItem
		var next bool
		for {
			item, next = iterFunc()
			if !next {
				break
			}

			var hash string
			sebakcommon.MustUnmarshalJSON(item.Value, &hash)
			chanWorker <- hash

			if hash == latestBlock.Hash {
				return
			}
		}
		closeFunc()
		close(chanWorker)
	}()

	var count uint64
end:
	for {
		select {
		case err := <-chanError:
			if err != nil {
				fmt.Println("error:", err)
				break end
			}
			count += 1
			if count == latestBlock.Height-blockHeight {
				break end
			}
			if count > 0 && count%1000 == 0 {
				fmt.Println("> remove count:", count)
			}
		}
	}

	// clean up confirmed
	iterFunc, closeFunc := st.GetIterator(
		sebakcommon.BlockPrefixConfirmed,
		sebakstorage.NewDefaultListOptions(
			true,
			nil,
			0,
		),
	)

	var countMissing int
	for {
		item, next := iterFunc()
		if !next {
			break
		}
		var hash string
		sebakcommon.MustUnmarshalJSON(item.Value, &hash)
		if _, err := sebakblock.GetBlock(st, hash); err != nil {
			countMissing += 1
			st.Remove(string(item.Key))
		}
	}
	closeFunc()

	fmt.Println("> remove missing", countMissing)
}
