package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	. "github.com/zitryss/aye-and-nay/internal/generator"
)

func TestBadgerTestSuite(t *testing.T) {
	suite.Run(t, &BadgerTestSuite{})
}

type BadgerTestSuite struct {
	suite.Suite
	base MemTestSuite
}

func (suite *BadgerTestSuite) SetupSuite() {
	if !*integration {
		suite.T().Skip()
	}
	suite.base = MemTestSuite{}
	suite.base.SetT(suite.T())
	ctx, cancel := context.WithCancel(context.Background())
	badger, err := NewBadger(DefaultBadgerConfig)
	require.NoError(suite.T(), err)
	suite.base.ctx = ctx
	suite.base.cancel = cancel
	suite.base.db = badger
}

func (suite *BadgerTestSuite) SetupTest() {
	err := suite.base.db.(*Badger).Reset()
	require.NoError(suite.T(), err)
}

func (suite *BadgerTestSuite) TearDownTest() {

}

func (suite *BadgerTestSuite) TearDownSuite() {
	err := suite.base.db.(*Badger).Reset()
	require.NoError(suite.T(), err)
	suite.base.cancel()
}

func (suite *BadgerTestSuite) TestBadgerAlbum() {
	suite.base.TestAlbum()
}

func (suite *BadgerTestSuite) TestBadgerCount() {
	suite.base.TestCount()
}

func (suite *BadgerTestSuite) TestBadgerImage() {
	suite.base.TestImage()
}

func (suite *BadgerTestSuite) TestBadgerVote() {
	suite.base.TestVote()
}

func (suite *BadgerTestSuite) TestBadgerSort() {
	suite.base.TestSort()
}

func (suite *BadgerTestSuite) TestBadgerRatings() {
	suite.base.TestRatings()
}

func (suite *BadgerTestSuite) TestBadgerDelete() {
	suite.base.TestDelete()
}

func (suite *BadgerTestSuite) TestBadgerLru() {
	id, ids := GenId()
	alb1 := suite.base.saveAlbum(id, ids)
	_ = suite.base.saveAlbum(id, ids)
	edgs, err := suite.base.db.GetEdges(suite.base.ctx, ids.Uint64(0))
	assert.NoError(suite.base.T(), err)
	assert.Equal(suite.base.T(), alb1.Edges, edgs)
}
