package main

import (
	"os"

	"github.com/nenad/couch/cmd"
	"github.com/nenad/couch/pkg/config"
	"github.com/nenad/couch/pkg/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	var dbName = os.Getenv("COUCH_DB")

	if dbName == "" {
		dbName = "couch.sqlite"
	}

	db, err := storage.NewCouchDatabase(dbName)
	if err != nil {
		logrus.Errorf("error while creating a database: %s", err)
		os.Exit(1)
	}

	store := &config.Store{DB: db}
	conf, err := config.InitConfiguration(store)
	if err != nil {
		if err := store.Save(conf); err != nil {
			logrus.Errorf("error while saving initial config: %s", err)
			os.Exit(1)
		}
	}

	rootCmd := cmd.NewCLI(conf, db)
	err = rootCmd.Execute()
	if err != nil {
		logrus.Errorf("error while executing command: %s", err)
		os.Exit(1)
	}
}
