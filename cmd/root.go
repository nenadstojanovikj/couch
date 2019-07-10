package cmd

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nenad/couch/pkg/config"
	"github.com/nenad/couch/pkg/notifications"
	"github.com/nenad/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCLI(conf config.Config, db *sql.DB) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "couch",
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		os.Exit(1)
	}()

	confStore := &config.Store{DB: db}
	repo := storage.NewMediaRepository(db)
	notifier := newNotifier(conf, db)
	rootCmd.AddCommand(NewAppCommand(conf, repo, notifier))
	rootCmd.AddCommand(NewAuthCommand(conf, confStore))

	return rootCmd
}

func newNotifier(conf config.Config, db *sql.DB) notifications.Notifier {
	if conf.TelegramBotToken == "" {
		return &notifications.NoopNotifier{}
	}

	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		logrus.Fatalf("error while creating Telegram Bot: %s", err)
	}

	client := notifications.NewTelegramClient(bot, db)
	go func() {
		if err := client.StartListener(); err != nil {
			logrus.Errorf("could not start Telegram listener: %s", err)
		}
	}()

	return client
}
