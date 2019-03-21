package newstorage

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type testNewStorage struct {
	suite.Suite
}

func TestNewStorage(t *testing.T) {
	suite.Run(t, new(testNewStorage))
}
