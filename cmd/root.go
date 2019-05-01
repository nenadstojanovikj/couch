package cmd

import (
	"database/sql"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func NewCLI(conf *config.Config, db *sql.DB) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "couch",
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		os.Exit(1)
	}()

	rootCmd.AddCommand(NewAppCommand(conf, db))
	rootCmd.AddCommand(NewAuthCommand(conf))

	return rootCmd
}
