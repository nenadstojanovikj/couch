package cmd

import (
	"context"
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/rd"
	"github.com/nenadstojanovikj/trakt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	cmd.AddCommand(&cobra.Command{
		Use: "trakt",
		Run: trakttv(config),
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
			conf.RealDebrid.ExpiresIn = int64(token.ExpiresIn)
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

func trakttv(conf *config.Config) func(command *cobra.Command, args []string) {
	clientID := "527bf0f0f3f6004266ad9a52a9c25a1f4547e09344b3b3abc467edd8cfbb2b73"
	secretID := "5c43a09e917e237ad7a0c6411e1e113176d0ac3810f81ebeb3f1f97534d6bf67"

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	cancChan := make(chan os.Signal, 1)
	signal.Notify(cancChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-cancChan
		cancel()
	}()

	return func(cmd *cobra.Command, args []string) {
		auth := trakt.AuthClient{
			ClientID:     clientID,
			ClientSecret: secretID,
			HttpClient:   http.DefaultClient,
		}

		code, err := auth.Code()
		if err != nil {
			logrus.Errorf("error while obtaining code: %s", err)
		}

		fmt.Printf("URL: %s\nCode: %s\n", code.VerificationURL, code.UserCode)
		fmt.Println("Press CTRL+C to cancel")

		result := <-auth.PollToken(ctx, code)

		if result.Err != nil {
			logrus.Errorf("error while obtaining trakt token: %s", result.Err)
			return
		}

		token := result.Token
		conf.Trakt.AccessToken = token.AccessToken
		conf.Trakt.RefreshToken = token.RefreshToken
		conf.Trakt.ClientSecret = secretID
		conf.Trakt.ClientID = clientID
		conf.Trakt.ObtainedAt = time.Unix(token.CreatedAt, 0)
		conf.Trakt.ExpiresIn = int64(token.ExpiresIn)
		conf.Trakt.TokenType = token.TokenType

		if err := conf.Save(); err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		logrus.Info("successfully obtained token")
	}
}
