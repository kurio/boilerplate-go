package mysql_test

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
)

func MigrateDB(db *sql.DB) (m *migrate.Migrate, err error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, err
	}
	m, err = migrate.NewWithDatabaseInstance("file://migrations/", "mysql", driver)
	if err != nil {
		return nil, err
	}
	err = m.Up()
	return
}
