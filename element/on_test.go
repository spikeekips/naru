package element

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/naru/common"
)

type testLevelDBItemEvent struct {
	suite.Suite
}

func (t *testLevelDBItemEvent) TestErrorReturn() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r)
		}
	}()

	ob := common.NewObservable("t")

	{ // no error
		ob.Sync("on-save", func(err *common.Error, a string) {
			// no error
		})

		var err *common.Error
		ob.Trigger("on-save", err, "findme")

		t.Nil(err)
	}

	{ // has error
		var errorCode uint = 10
		ob.Sync("on-save", func(err *common.Error, a string) {
			e := common.NewError(errorCode, "has error?")
			*err = *e
		})

		var err common.Error
		ob.Trigger("on-save", &err, "findme")
		t.NotNil(err)
		t.Equal(errorCode, err.Code())
	}
}

func (t *testLevelDBItemEvent) TestWrap() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			panic(r)
		}
	}()

	ob := common.NewObservable("t")

	{
		var errorCode uint = 10
		onF := func(a string) error {
			return common.NewError(errorCode, "has error?")
		}

		ob.Sync("on-save", SyncWithError(onF))

		var err common.Error
		ob.Trigger("on-save", &err, "findme")
		t.Equal(errorCode, err.Code())
	}
}

func TestLevelDBItemEvent(t *testing.T) {
	suite.Run(t, new(testLevelDBItemEvent))
}
