package sebak

import (
	"bytes"
	"net/http"
	"sync"

	jsonrpc "github.com/gorilla/rpc/json"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakrunner "boscoin.io/sebak/lib/node/runner"
	sebakstorage "boscoin.io/sebak/lib/storage"
)

type StorageProvider interface {
	Open() error
	Close() error
	New() StorageProvider
	Has(string) (bool, error)
	Get(string) ([]byte, error)
	GetIterator(string, sebakstorage.ListOptions) (uint64, []sebakstorage.IterItem, error) // (limit, items, error)
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

func (j *JSONRPCStorageProvider) GetIterator(prefix string, options sebakstorage.ListOptions) (uint64, []sebakstorage.IterItem, error) {
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

	var result sebakrunner.DBGetIteratorResult
	err := j.request("DB.GetIterator", &args, &result)
	if err != nil {
		return 0, nil, err
	}

	return result.Limit, result.Items, nil
}
