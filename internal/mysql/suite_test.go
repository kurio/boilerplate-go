package mysql_test

import (
	"database/sql"
	"os"

	"github.com/golang-migrate/migrate"

	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/suite"
)

type mysqlSuite struct {
	suite.Suite
	DB *sql.DB
	mg *migrate.Migrate
}

func (m *mysqlSuite) SetupSuite() {
	dsnDB := os.Getenv("MYSQL_TEST")
	if dsnDB == "" {
		dsnDB = "kurio:supersecret@tcp(localhost:3306)/myDB?parseTime=1&loc=Asia%2FJakarta&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	}

	db, err := sql.Open("mysql", dsnDB)
	require.NoError(m.T(), err)
	require.NotNil(m.T(), db)
	m.mg, err = MigrateDB(db)
	require.NoError(m.T(), err)
	m.DB = db
}

func (m *mysqlSuite) TearDownSuite() {
	require.NoError(m.T(), m.mg.Drop())
	require.NoError(m.T(), m.DB.Close())
}
