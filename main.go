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

	err = conf.Load()
	if err != nil {
		if err := conf.Save(); err != nil {
			logrus.Errorf("error while saving initial config: %s", err)
			os.Exit(1)
		}
	}

	rootCmd := cmd.NewCLI(&conf, db)
	err = rootCmd.Execute()
	if err != nil {
		logrus.Errorf("error while executing command: %s", err)
		os.Exit(1)
	}

}
