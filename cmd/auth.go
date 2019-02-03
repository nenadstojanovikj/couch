package cmd

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/rd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"time"
)

func NewAuthCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "auth",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: auth [realdebrid,trakt]")
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use: "realdebrid",
		Run: realdebrid(config),
	})

	return cmd
}

func realdebrid(conf *config.Config) func(command *cobra.Command, args []string) {
	rdClientId := "X245A4XAIBGVM"
	return func(cmd *cobra.Command, args []string) {
		auth := rd.NewAuthClient(http.DefaultClient)
		creds, err := auth.StartAuthentication(rdClientId)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Direct Url: %s\n\n%s\nCode: %s\n", creds.DirectVerificationURL, creds.VerificationURL, creds.UserCode)
		tries := creds.ExpiresIn / creds.Interval
		var token rd.Token

		for i := 1; i <= tries; i++ {
			time.Sleep(time.Duration(creds.Interval) * time.Second)
			secrets, err := auth.ObtainSecret(creds.DeviceCode, rdClientId)
			if err != nil {
				fmt.Printf("Still not verified, retrying (%d/%d) - Ctrl+C to cancel\n", i, tries)
				continue
			}
			token, err = auth.ObtainAccessToken(secrets.ClientID, secrets.ClientSecret, creds.DeviceCode)
			if err != nil {
				fmt.Printf("Error while obtaining token: %s", err)
				os.Exit(1)
			}

			conf.RealDebrid.AccessToken = token.AccessToken
			conf.RealDebrid.RefreshToken = token.RefreshToken
			conf.RealDebrid.ClientSecret = secrets.ClientSecret
			conf.RealDebrid.ClientID = secrets.ClientID
			conf.RealDebrid.ObtainedAt = token.ObtainedAt
			conf.RealDebrid.ExpiresIn = token.ExpiresIn
			conf.RealDebrid.TokenType = token.TokenType

			if err := conf.Save(); err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
			break
		}

		if token.AccessToken != "" {
			fmt.Println("Successfully obtained token.")
		} else {
			fmt.Println("Failed to store token.")
		}

	}
}
