package sebak

import (
	"fmt"
	"testing"

	sebakcommon "boscoin.io/sebak/lib/common"
	sebakerrors "boscoin.io/sebak/lib/errors"
	sebakstorage "boscoin.io/sebak/lib/storage"
	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/naru/storage"
)

type FakeSEBAKStorageProvider struct {
	OpenError     error
	CloseError    error
	HasError      error
	GetError      error
	IteratorError error

	items    []sebakstorage.IterItem
	itemsMap map[string]sebakstorage.IterItem

	Limit uint64
}

func newFakeSEBAKStorageProvider() *FakeSEBAKStorageProvider {
	return &FakeSEBAKStorageProvider{
		items:    []sebakstorage.IterItem{},
		itemsMap: map[string]sebakstorage.IterItem{},
	}
}

func (f *FakeSEBAKStorageProvider) Items() []sebakstorage.IterItem {
	return f.items
}

func (f *FakeSEBAKStorageProvider) SetItems(items []sebakstorage.IterItem) {
	f.items = []sebakstorage.IterItem{}
	f.itemsMap = map[string]sebakstorage.IterItem{}

	for n, item := range items {
		item.N = uint64(n)
		f.itemsMap[string(item.Key)] = item
		f.items = append(f.items, item)
	}
}

func (f *FakeSEBAKStorageProvider) Open() error {
	return f.OpenError
}

func (f *FakeSEBAKStorageProvider) Close() error {
	return f.CloseError
}

func (f *FakeSEBAKStorageProvider) New() StorageProvider {
	return newFakeSEBAKStorageProvider()
}

func (f *FakeSEBAKStorageProvider) Has(key string) (bool, error) {
	if f.HasError != nil {
		return false, f.HasError
	}

	_, found := f.itemsMap[key]
	return found, nil
}

func (f *FakeSEBAKStorageProvider) Get(key string) ([]byte, error) {
	if f.GetError != nil {
		return nil, f.GetError
	}

	item, found := f.itemsMap[key]
	if !found {
		return nil, sebakerrors.StorageRecordDoesNotExist
	}

	return item.Value, nil
}

func (f *FakeSEBAKStorageProvider) Iterator(prefix string, options sebakstorage.ListOptions) (uint64, []sebakstorage.IterItem, error) {
	if f.IteratorError != nil {
		return 0, nil, f.IteratorError
	}

	var limit uint64
	if f.Limit < 1 {
		limit = options.Limit()
	} else {
		limit = f.Limit
	}

	source := make([]sebakstorage.IterItem, len(f.items))
	copy(source, f.items)

	if options.Reverse() {
		for i, j := 0, len(source)-1; i < j; i, j = i+1, j-1 {
			source[i], source[j] = source[j], source[i]
		}
	}

	var items []sebakstorage.IterItem
	if len(options.Cursor()) < 1 {
		items = source[:limit]
	} else {
		var found bool
		for _, item := range source {
			if string(item.Key) == string(options.Cursor()) {
				found = true
				continue
			}
			if !found {
				continue
			}
			items = append(items, item)
			if len(items) == int(options.Limit()) {
				break
			}
		}
	}

	return limit, items, nil
}

type testSuiteSEBAKStorageProvider struct {
	suite.Suite
}

func (t *testSuiteSEBAKStorageProvider) makeItems(n int) (items []sebakstorage.IterItem) {
	for i := 0; i < n; i++ {
		v, _ := storage.Serialize(sebakcommon.GetUniqueIDFromUUID())
		items = append(
			items,
			sebakstorage.IterItem{
				N:     uint64(i),
				Key:   []byte(sebakcommon.GetUniqueIDFromUUID()),
				Value: v,
			},
		)
	}

	return
}

func (t *testSuiteSEBAKStorageProvider) TestHas() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)
	for _, item := range p.Items() {
		found, err := st.Has(string(item.Key))
		t.NoError(err)
		t.True(found)
	}
}

func (t *testSuiteSEBAKStorageProvider) TestHasUnknownKey() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)
	found, err := st.Has(sebakcommon.GetUniqueIDFromUUID())
	t.NoError(err)
	t.False(found)
}

func (t *testSuiteSEBAKStorageProvider) TestHasError() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(2))
	p.GetError = fmt.Errorf("something wrong")

	st := NewStorage(p)

	var returned string
	_, err := st.Get(string(p.Items()[1].Key), &returned)
	t.Equal(p.GetError, err)
}

func (t *testSuiteSEBAKStorageProvider) TestGet() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)
	for _, item := range p.Items() {
		var returned string
		var expected string
		storage.Deserialize(item.Value, &expected)

		_, err := st.Get(string(item.Key), &returned)
		t.NoError(err)

		t.Equal(expected, returned)
	}
}

func (t *testSuiteSEBAKStorageProvider) TestGetUnknownKey() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)

	var returned string
	_, err := st.Get(sebakcommon.GetUniqueIDFromUUID(), &returned)
	t.Equal(err, sebakerrors.StorageRecordDoesNotExist)
}

func (t *testSuiteSEBAKStorageProvider) TestGetError() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(2))
	p.GetError = fmt.Errorf("something wrong")

	st := NewStorage(p)

	var returned string
	_, err := st.Get(string(p.Items()[1].Key), &returned)
	t.Equal(p.GetError, err)
}

func (t *testSuiteSEBAKStorageProvider) TestIterator() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)

	iterFunc, closeFunc := st.Iterator("", sebakstorage.NewDefaultListOptions(false, nil, uint64(len(p.Items()))))
	defer closeFunc()

	var items []sebakstorage.IterItem
	for {
		item, next := iterFunc()
		if !next {
			break
		}

		items = append(items, item)
	}

	t.Equal(len(p.Items()), len(items))

	for n, item := range items {
		expected := p.Items()[n]
		t.Equal(string(expected.Key), string(item.Key))
		t.Equal(string(expected.Value), string(item.Value))
	}
}

func (t *testSuiteSEBAKStorageProvider) TestIteratorCloseFunc() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)

	iterFunc, closeFunc := st.Iterator("", sebakstorage.NewDefaultListOptions(false, nil, uint64(len(p.Items()))))

	iterFunc()
	closeFunc()
	item, next := iterFunc()

	t.Empty(item)
	t.False(next)
}

func (t *testSuiteSEBAKStorageProvider) TestIteratorReverse() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)

	iterFunc, closeFunc := st.Iterator("", sebakstorage.NewDefaultListOptions(true, nil, uint64(len(p.Items()))))
	defer closeFunc()

	var items []sebakstorage.IterItem
	for {
		item, next := iterFunc()
		if !next {
			break
		}

		items = append(items, item)
	}

	t.Equal(len(p.Items()), len(items))

	for n, item := range items {
		expected := p.Items()[len(p.Items())-n-1]
		t.Equal(string(expected.Key), string(item.Key))
		t.Equal(string(expected.Value), string(item.Value))
	}
}

func (t *testSuiteSEBAKStorageProvider) TestIteratorError() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))
	p.IteratorError = fmt.Errorf("something wrong")

	st := NewStorage(p)

	iterFunc, closeFunc := st.Iterator("", sebakstorage.NewDefaultListOptions(true, nil, uint64(len(p.Items()))))
	defer closeFunc()

	item, next := iterFunc()
	t.Empty(item)
	t.False(next)
}

func (t *testSuiteSEBAKStorageProvider) TestIteratorLimit() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)

	limit := 3
	iterFunc, closeFunc := st.Iterator("", sebakstorage.NewDefaultListOptions(false, nil, uint64(limit)))
	defer closeFunc()

	var items []sebakstorage.IterItem
	for {
		item, next := iterFunc()
		if !next {
			break
		}

		items = append(items, item)
	}

	t.Equal(limit, len(items))

	for n, item := range items {
		expected := p.Items()[n]
		t.Equal(string(expected.Key), string(item.Key))
		t.Equal(string(expected.Value), string(item.Value))
	}
}

func (t *testSuiteSEBAKStorageProvider) TestIteratorLimitReverse() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))

	st := NewStorage(p)

	limit := 3
	iterFunc, closeFunc := st.Iterator("", sebakstorage.NewDefaultListOptions(true, nil, uint64(limit)))
	defer closeFunc()

	var items []sebakstorage.IterItem
	for {
		item, next := iterFunc()
		if !next {
			break
		}

		items = append(items, item)
	}

	t.Equal(limit, len(items))

	for n, item := range items {
		expected := p.Items()[len(p.Items())-n-1]
		t.Equal(string(expected.Key), string(item.Key))
		t.Equal(string(expected.Value), string(item.Value))
	}
}

func (t *testSuiteSEBAKStorageProvider) TestIteratorOverLimit() {
	p := newFakeSEBAKStorageProvider()
	p.SetItems(t.makeItems(5))
	p.Limit = 3

	st := NewStorage(p)

	iterFunc, closeFunc := st.Iterator("", sebakstorage.NewDefaultListOptions(false, nil, 1000))
	defer closeFunc()

	var items []sebakstorage.IterItem
	for {
		item, next := iterFunc()
		if !next {
			break
		}

		items = append(items, item)
	}

	t.Equal(len(p.Items()), len(items))

	for n, item := range items {
		expected := p.Items()[n]
		t.Equal(string(expected.Key), string(item.Key))
		t.Equal(string(expected.Value), string(item.Value))
	}
}

func TestSEBAKStorageProvider(t *testing.T) {
	suite.Run(t, new(testSuiteSEBAKStorageProvider))
}
