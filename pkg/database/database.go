package database

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Connect(path string) (*sqlx.DB, error) {
	db, err := getConnection(path)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getConnection(path string) (*sqlx.DB, error) {
	if !strings.HasSuffix(path, ".db") {
		return nil, fmt.Errorf("Please specifiy a path to a sql lite database")
	}

	_, err := os.Stat(path)
	notExist := errors.Is(err, os.ErrNotExist)

	if notExist {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %v", err)
	}

	if notExist {
		err := createTable(db)
		if err != nil {
			return nil, fmt.Errorf("Error creating table: %v", err)
		}
	}

	return db, nil
}

func createTable(db *sqlx.DB) error {
	query := `CREATE TABLE products (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"name" TEXT,
		"price" REAL,
		"available" INTEGER,
		"link" TEXT,
		"inserted_at" INTEGER
	  );`

	_, err := db.Exec(query)
	return err
}
