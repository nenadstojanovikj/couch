package cmd

import (
	"database/sql"
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/spf13/cobra"
)

func NewCLI(conf *config.Config, db *sql.DB) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "couch",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: couch auth")
		},
	}

	repo := storage.NewMediaRepository(db)

	rootCmd.AddCommand(NewAppCommand(conf, repo))
	rootCmd.AddCommand(NewAuthCommand(conf))

	return rootCmd
}
