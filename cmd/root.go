package cmd

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/spf13/cobra"
	"os"
)

func NewCLI(conf *config.Config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "couch",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: couch auth")
		},
	}

	db, err := storage.NewCouchDatabase("couch.sqlite")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	repo := storage.NewMediaRepository(db)

	rootCmd.AddCommand(NewAppCommand(conf, repo))
	rootCmd.AddCommand(NewAuthCommand(conf))

	return rootCmd
}
