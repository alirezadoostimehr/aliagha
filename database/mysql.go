package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	msql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/go-sql-driver/mysql"

	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateDB(username, password, host, port, dbname, address, job string) error {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		username, password, host, port)

	dbStart, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	_, err = dbStart.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", "aliagha"))
	if err != nil {
		panic(err)
	}

	dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		username, password, host, port, dbname)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", address),
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}

	if job == "up" {
		err = m.Up()
	} else {
		err = m.Down()
	}
	return err
}

func InitDB(username, password, host, name, dbname string, port int) (*gorm.DB, error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		username, password, host, port, name)
	db, err := gorm.Open(msql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
