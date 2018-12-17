package cmd

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/spf13/cobra"
)

func NewCLI(conf *config.Config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "couch",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: couch auth")
		},
	}

	rootCmd.AddCommand(NewAuthCommand(conf))

	return rootCmd
}
