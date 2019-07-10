package config

import (
	"database/sql"
	"encoding/json"
)

type (
	Saver interface {
		// Save stores the structure somewhere
		Save(v interface{}) error
	}

	Loader interface {
		// Load returns store from somewhere
		Load() (interface{}, error)
	}

	SaveLoader interface {
		Saver
		Loader
	}

	Store struct {
		DB *sql.DB
	}
)

func (s *Store) Save(v interface{}) error {
	// Indenting so it's human readable for easier inspection
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	_, err = s.DB.Exec("UPDATE config SET config = ?", b)
	return err
}

func (s *Store) Load() (interface{}, error) {
	row := s.DB.QueryRow("SELECT config FROM config LIMIT 1;")

	var j []byte
	err := row.Scan(&j)
	if err != nil {
		return nil, err
	}

	var c Config
	err = json.Unmarshal(j, &c)
	return c, err
}
