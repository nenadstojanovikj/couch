package cmd

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/nenadstojanovikj/rd"
	"github.com/nenadstojanovikj/showrss-go"
	"github.com/spf13/cobra"
	"net/http"
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
	repo := storage.NewMediaItemRepository(db)

	debrid := rd.NewRealDebrid(createToken(&conf.RealDebrid), http.DefaultClient, rd.AutoRefresh)
	feed := showrss.NewClient(http.DefaultClient)

	rootCmd.AddCommand(NewAuthCommand(conf))
	rootCmd.AddCommand(NewServerCommand(conf))
	rootCmd.AddCommand(NewFetchCommand(conf.ShowRss.PersonalFeed, repo, feed, debrid))

	return rootCmd
}

func createToken(conf *config.AuthConfig) rd.Token {
	return rd.Token{
		AccessToken:  conf.AccessToken,
		TokenType:    conf.TokenType,
		ExpiresIn:    conf.ExpiresIn,
		ObtainedAt:   conf.ObtainedAt,
		RefreshToken: conf.RefreshToken,
	}
}
