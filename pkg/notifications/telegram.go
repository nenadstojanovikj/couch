package notifications

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nenad/couch/pkg/media"
	"github.com/sirupsen/logrus"
)

type Notifier interface {
	OnQueued(item media.SearchItem) error
	OnFinish(item media.SearchItem) error
}

type Telegram struct {
	bot *tgbotapi.BotAPI
	db  *sql.DB
}

func NewTelegramClient(bot *tgbotapi.BotAPI, db *sql.DB) *Telegram {
	return &Telegram{
		bot: bot,
		db:  db,
	}
}
func (t *Telegram) OnQueued(item media.SearchItem) error {
	for _, s := range t.GetSubscribedChats() {
		if _, err := t.bot.Send(tgbotapi.NewMessage(s, fmt.Sprintf("%q was queued for downloading.", item.Term))); err != nil {
			return err
		}
	}
	return nil
}

func (t *Telegram) OnFinish(item media.SearchItem) error {
	for _, s := range t.GetSubscribedChats() {
		if _, err := t.bot.Send(tgbotapi.NewMessage(s, fmt.Sprintf("%q was downloaded.", item.Term))); err != nil {
			return err
		}
	}
	return nil
}

func (t *Telegram) GetSubscribedChats() (ids []int64) {
	rows, err := t.db.Query("SELECT id FROM telegram")
	if err != nil {
		return ids
	}
	var id int64
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			break
		}
		ids = append(ids, id)
	}

	return ids
}

func (t *Telegram) UpdateSubscribers(text string) error {
	for _, id := range t.GetSubscribedChats() {
		_, err := t.bot.Send(tgbotapi.NewMessage(id, text))
		if err != nil {
			return fmt.Errorf("unable to send update message: %s", err)
		}
	}

	return nil
}

func (t *Telegram) RegisterChat(id int64) error {
	_, err := t.db.Exec("INSERT INTO telegram (id) VALUES (?)", id)
	if err != nil {
		return fmt.Errorf("error while registering Telegram chat: %s", err)
	}
	return nil
}

func (t *Telegram) UnregisterChat(id int64) error {
	_, err := t.db.Exec("DELETE FROM telegram WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error while unregistering Telegram chat: %s", err)
	}
	return nil
}

func (t *Telegram) StartListener() error {
	u := tgbotapi.NewUpdate(0)
	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("error while trying to get updates channel: %s", err)
	}

	logrus.Infof("started listener")
	for update := range updates {
		logrus.Infof("received message")
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "subscribe":
			if err := t.RegisterChat(update.Message.Chat.ID); err != nil {
				logrus.Warnf("error while registering chat: %s", err)
			}
		case "unsubscribe":
			if err := t.UnregisterChat(update.Message.Chat.ID); err != nil {
				logrus.Warnf("error while unregistering chat: %s", err)
			}
		default:
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You have been successfully %sd", update.Message.Command()))
		msg.ReplyToMessageID = update.Message.MessageID

		_, err := t.bot.Send(msg)
		if err != nil {
			logrus.Warnf("error while sending a message: %s", err)
		}
	}
	logrus.Infof("ended listener")

	return nil
}
