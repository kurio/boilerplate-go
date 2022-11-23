package mysql_test

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MySQLSuite ...
type MySQLSuite struct {
	suite.Suite
	DBConn *sql.DB
	mg     *migrate.Migrate
}

func (s *MySQLSuite) SetupSuite() {
	var err error

	dsnDB := os.Getenv("MYSQL_TEST")
	if dsnDB == "" {
		dsnDB = "user:pass@tcp(localhost:3306)/myDB?parseTime=1&loc=Asia%2FJakarta&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	}

	db, err := sql.Open("mysql", dsnDB)
	require.NoError(s.T(), err)

	require.True(s.T(), s.tryPing())

	s.DBConn = db

	s.migrateDatabase()
	require.NoError(s.T(), err)
}

func (s *MySQLSuite) tryPing() bool {
	s.T().Helper()

	maxWaitTime := 1000 * time.Millisecond
	for err, t := s.DBConn.Ping(), 10*time.Millisecond; t < maxWaitTime; {
		if err == nil {
			return true
		}

		time.Sleep(t)

		err = s.DBConn.Ping()
		t = 2 * t
	}

	return false
}

func (s *MySQLSuite) TearDownSuite() {
	require.NoError(s.T(), s.mg.Drop())
	require.NoError(s.T(), s.DBConn.Close())
}

func (s *MySQLSuite) migrateDatabase() {
	t := s.T()
	t.Helper()

	driver, err := mysql.WithInstance(s.DBConn, &mysql.Config{})
	require.NoError(t, err)

	m, err := migrate.NewWithDatabaseInstance("file://migrations/", "mysql", driver)
	require.NoError(t, err)

	require.NoError(t, m.Drop())
	require.NoError(t, m.Up())

	s.mg = m
}
