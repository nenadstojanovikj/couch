package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nenadstojanovikj/couch/resources"
	"github.com/sirupsen/logrus"
)

const ISO8601 string = "2006-01-02 03:04:05.000"

func NewCouchDatabase(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filename+"?cache=shared&_fk=true&_journal=WAL")
	db.SetMaxOpenConns(1)
	if err != nil {
		return nil, err
	}

	version := getCurrentVersion(db)
	newMigrations := resources.Migrations()[version:]

	// Apply the migrations one by one
	for i, mig := range newMigrations {
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(mig)
		if err != nil {
			return nil, err
		}

		version++
		_, err = tx.Exec("UPDATE version SET version = ?", version)
		if err != nil {
			return nil, err
		}

		err = tx.Commit()
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return nil, err
			}
			return nil, err
		}
		logrus.Debugf("applied migration %d successfully", i)
	}

	return db, err
}

func getCurrentVersion(db *sql.DB) (version int) {
	r := db.QueryRow("SELECT version FROM version")
	if err := r.Scan(&version); err != nil {
		return 0
	} else {
		return version
	}
}
