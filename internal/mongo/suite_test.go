package mongo_test

import (
	"context"
	"os"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoSuite struct {
	suite.Suite
	database *mongo.Database
}

func (s *mongoSuite) SetupSuite() {
	mongoDSN := os.Getenv("MONGO_TEST_URI")
	if mongoDSN == "" {
		mongoDSN = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_TEST_DB")
	if dbName == "" {
		dbName = "testDB"
	}

	client, err := mongo.Connect(context.Background(),
		options.Client().
			SetConnectTimeout(2*time.Second).
			SetServerSelectionTimeout(3*time.Second).
			ApplyURI(mongoDSN),
	)
	require.NoError(s.T(), err)

	err = client.Ping(context.Background(), nil)
	require.NoError(s.T(), err)

	s.database = client.Database(dbName)
}

func (s *mongoSuite) TearDownSuite() {
	err := s.database.Client().Disconnect(context.Background())
	require.NoError(s.T(), err)
}
