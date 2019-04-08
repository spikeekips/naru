package sebak

import (
	"bytes"
	"net/http"
	"sync"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakrunner "boscoin.io/sebak/lib/node/runner"
	jsonrpc "github.com/gorilla/rpc/json"

	"github.com/spikeekips/naru/storage"
	leveldbstorage "github.com/spikeekips/naru/storage/backend/leveldb"
)

type StorageProvider interface {
	Open() error
	Close() error
	New() StorageProvider
	Has(string) (bool, error)
	Get(string) ([]byte, error)
	GetIterator(string, storage.ListOptions) (func() (storage.IterItem, bool), func())
}

type JSONRPCStorageProvider struct {
	sync.RWMutex
	endpoint *sebakcommon.Endpoint
	snapshot string
}

func NewJSONRPCStorageProvider(endpoint *sebakcommon.Endpoint) *JSONRPCStorageProvider {
	return &JSONRPCStorageProvider{
		endpoint: endpoint,
	}
}

func (j *JSONRPCStorageProvider) setSnapshot(snapshot string) {
	j.Lock()
	defer j.Unlock()

	j.snapshot = snapshot
}

func (j *JSONRPCStorageProvider) getSnapshot() string {
	j.RLock()
	defer j.RUnlock()

	return j.snapshot
}

func (j *JSONRPCStorageProvider) request(method string, args interface{}, result interface{}) error {
	message, err := jsonrpc.EncodeClientRequest(method, &args)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", j.endpoint.String(), bytes.NewBuffer(message))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return jsonrpc.DecodeClientResponse(resp.Body, result)
}

func (j *JSONRPCStorageProvider) Open() error {
	if len(j.getSnapshot()) > 0 {
		return ProviderNotClosedError
	}

	var result sebakrunner.DBOpenSnapshotResult
	args := &sebakrunner.DBOpenSnapshot{}
	if err := j.request("DB.OpenSnapshot", args, &result); err != nil {
		return err
	}
	j.setSnapshot(result.Snapshot)

	return nil
}

func (j *JSONRPCStorageProvider) Close() error {
	if len(j.getSnapshot()) < 1 {
		return ProviderNotOpenedError
	}

	var result sebakrunner.DBReleaseSnapshotResult
	args := &sebakrunner.DBReleaseSnapshot{Snapshot: j.getSnapshot()}
	if err := j.request("DB.ReleaseSnapshot", args, &result); err != nil {
		return err
	}

	if !bool(result) {
		log.Warn("`JSONRPCStorageProvider` was not cleanly released")
	}

	defer func() {
		j.setSnapshot("")
	}()

	return nil
}

func (j *JSONRPCStorageProvider) New() StorageProvider {
	return NewJSONRPCStorageProvider(j.endpoint)
}

func (j *JSONRPCStorageProvider) Has(key string) (bool, error) {
	if len(j.getSnapshot()) < 1 {
		return false, ProviderNotOpenedError
	}

	var result sebakrunner.DBHasResult
	err := j.request("DB.Has", &sebakrunner.DBHasArgs{Snapshot: j.getSnapshot(), Key: key}, &result)
	if err != nil {
		return false, err
	}

	return bool(result), nil
}

func (j *JSONRPCStorageProvider) Get(key string) ([]byte, error) {
	if len(j.getSnapshot()) < 1 {
		return nil, ProviderNotOpenedError
	}

	var result sebakrunner.DBGetResult
	err := j.request("DB.Get", &sebakrunner.DBGetArgs{Snapshot: j.getSnapshot(), Key: key}, &result)
	if err != nil {
		return nil, err
	}

	return result.Value, nil
}

func (j *JSONRPCStorageProvider) GetIterator0(prefix string, options storage.ListOptions) (uint64, []storage.IterItem, error) {
	if len(j.getSnapshot()) < 1 {
		return 0, nil, ProviderNotOpenedError
	}

	args := sebakrunner.DBGetIteratorArgs{
		Snapshot: j.getSnapshot(),
		Prefix:   prefix,
		Options: sebakrunner.GetIteratorOptions{
			Reverse: options.Reverse(),
			Cursor:  options.Cursor(),
			Limit:   options.Limit(),
		},
	}

	var result storage.SEBAKDBGetIteratorResult
	err := j.request("DB.GetIterator", &args, &result)
	if err != nil {
		return 0, nil, err
	}

	return result.Limit, result.Items, nil
}

func (j *JSONRPCStorageProvider) GetIterator(prefix string, options storage.ListOptions) (func() (storage.IterItem, bool), func()) {
	nullIterFunc := func() (storage.IterItem, bool) {
		return storage.IterItem{}, false
	}
	nullCloseFunc := func() {}

	if len(j.getSnapshot()) < 1 {
		return nullIterFunc, nullCloseFunc
	}

	dbGetIterator := func(options storage.ListOptions) (uint64, []storage.IterItem, error) {
		if len(j.getSnapshot()) < 1 {
			return 0, nil, ProviderNotOpenedError
		}

		args := sebakrunner.DBGetIteratorArgs{
			Snapshot: j.getSnapshot(),
			Prefix:   prefix,
			Options: sebakrunner.GetIteratorOptions{
				Reverse: options.Reverse(),
				Cursor:  options.Cursor(),
				Limit:   options.Limit(),
			},
		}

		var result storage.SEBAKDBGetIteratorResult
		err := j.request("DB.GetIterator", &args, &result)
		if err != nil {
			return 0, nil, err
		}

		return result.Limit, result.Items, nil
	}

	var cursor []byte = options.Cursor()
	var limit uint64
	var items []storage.IterItem
	var err error
	var closed bool

	var all int
	var n int
	var iterFunc func() (storage.IterItem, bool)
	iterFunc = func() (storage.IterItem, bool) {
		if closed {
			return storage.IterItem{}, false
		}

		if options.Limit() > 0 && uint64(all) >= options.Limit() {
			closed = true
			return storage.IterItem{}, false
		}

		if items == nil {
			options.SetCursor(cursor)
			limit, items, err = dbGetIterator(options)
			n = 0
		}

		if err != nil {
			log.Error("failed GetIterator", "error", err)
			return storage.IterItem{}, false
		}
		if len(items) == 0 {
			return storage.IterItem{}, false
		}
		if len(items) >= n+1 {
			defer func() {
				n += 1
				all += 1
			}()
			return items[n], true
		}
		if int(limit) > len(items) {
			return storage.IterItem{}, false
		}

		items = nil

		return iterFunc()
	}

	return iterFunc, func() {
		closed = true
	}
}

type LocalStorageProvider struct {
	s *leveldbstorage.Storage
}

func NewLocalStorageProvider(s *leveldbstorage.Storage) *LocalStorageProvider {
	return &LocalStorageProvider{s: s}
}

func (j *LocalStorageProvider) Open() error {
	return nil
}

func (j *LocalStorageProvider) Close() error {
	return nil
}

func (j *LocalStorageProvider) New() StorageProvider {
	return j
}

func (j *LocalStorageProvider) Has(key string) (bool, error) {
	return j.s.Has(key)
}

func (j *LocalStorageProvider) Get(key string) ([]byte, error) {
	return j.s.GetRaw(key)
}

func (j *LocalStorageProvider) GetIterator(prefix string, options storage.ListOptions) (func() (storage.IterItem, bool), func()) {
	return j.s.IteratorRaw(prefix, options)
}
