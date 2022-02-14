package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	. "github.com/zitryss/aye-and-nay/internal/generator"
)

func TestMongoTestSuite(t *testing.T) {
	suite.Run(t, &MongoTestSuite{})
}

type MongoTestSuite struct {
	suite.Suite
	base MemTestSuite
}

func (suite *MongoTestSuite) SetupSuite() {
	if !*integration {
		suite.T().Skip()
	}
	suite.base = MemTestSuite{}
	suite.base.SetT(suite.T())
	ctx, cancel := context.WithCancel(context.Background())
	mongo, err := NewMongo(ctx, DefaultMongoConfig)
	require.NoError(suite.T(), err)
	suite.base.ctx = ctx
	suite.base.cancel = cancel
	suite.base.db = mongo
}

func (suite *MongoTestSuite) SetupTest() {
	err := suite.base.db.(*Mongo).Reset()
	require.NoError(suite.T(), err)
}

func (suite *MongoTestSuite) TearDownTest() {

}

func (suite *MongoTestSuite) TearDownSuite() {
	err := suite.base.db.(*Mongo).Reset()
	require.NoError(suite.T(), err)
	err = suite.base.db.(*Mongo).Close(suite.base.ctx)
	require.NoError(suite.T(), err)
	suite.base.cancel()
}

func (suite *MongoTestSuite) TestMongoAlbum() {
	suite.base.TestAlbum()
}

func (suite *MongoTestSuite) TestMongoCount() {
	suite.base.TestCount()
}

func (suite *MongoTestSuite) TestMongoImage() {
	suite.base.TestImage()
}

func (suite *MongoTestSuite) TestMongoVote() {
	suite.base.TestVote()
}

func (suite *MongoTestSuite) TestMongoSort() {
	suite.base.TestSort()
}

func (suite *MongoTestSuite) TestMongoRatings() {
	suite.base.TestRatings()
}

func (suite *MongoTestSuite) TestMongoDelete() {
	suite.base.TestDelete()
}

func (suite *MongoTestSuite) TestMongoLru() {
	id, ids := GenId()
	alb1 := suite.base.saveAlbum(id, ids)
	_ = suite.base.saveAlbum(id, ids)
	edgs, err := suite.base.db.GetEdges(suite.base.ctx, ids.Uint64(0))
	assert.NoError(suite.base.T(), err)
	assert.Equal(suite.base.T(), alb1.Edges, edgs)
}
