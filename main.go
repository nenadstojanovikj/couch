package main

import (
	"github.com/nenadstojanovikj/couch/cmd"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"os"
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

	conf := config.NewConfiguration(db)
	if err := conf.Load(); err != nil {
		logrus.Errorf("error while loading configuration: %s", err)
		os.Exit(1)
	}

	rootCmd := cmd.NewCLI(&conf, db)
	err = rootCmd.Execute()
	if err != nil {
		logrus.Errorf("error while executing command: %s", err)
		os.Exit(1)
	}

}
