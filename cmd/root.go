package cmd

import (
	"database/sql"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/spf13/cobra"
)

func NewCLI(conf *config.Config, db *sql.DB) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "couch",
	}

	rootCmd.AddCommand(NewAppCommand(conf, db))
	rootCmd.AddCommand(NewAuthCommand(conf))

	return rootCmd
}
