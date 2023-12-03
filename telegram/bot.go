package telegram

import (
	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const telegramAlert = "telegram_alert"

type Bot struct {
	bot       *tgbotapi.BotAPI
	appLogger *logger.Logger
	cfg       *config.Config
	strg      storage.StorageI
}

func NewBot(bot *tgbotapi.BotAPI, appLogger *logger.Logger, cfg *config.Config, strg storage.StorageI) *Bot {
	return &Bot{
		bot:       bot,
		appLogger: appLogger,
		cfg:       cfg,
		strg:      strg,
	}
}

func (b *Bot) Start() error {
	b.appLogger.Infof("Authired on Account: %v", b.bot.Self.UserName)

	updates := b.initUpdatesChannel()

	b.handleUpdates(updates)

	return nil
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		// Check if the update contains the start command
		if update.Message == nil {
			continue
		}
		if update.Message.Command() == "start" {
			if b.handleCommandsStart(update) {
				continue
			}
		}
		// TODO: handle language keyboard and so on
	}
}

func (b *Bot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.bot.GetUpdatesChan(u)
}

func (b *Bot) handleSendTextMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)

	b.bot.Send(msg)
}

func (b *Bot) handleSuccessfullyLinkedAccount(update *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🌟 Fantastic! Your CertAlert account is now linked for SSL alerts in CertAlert's Telegram Bot. 🌐🛡️ Get ready to receive domain updates and SSL expiration alerts to keep your domains secure! 🔒✨\n\nPlease select your preferred language.")

	msg.ReplyMarkup = chooseLanguageStart

	if _, err := b.bot.Send(msg); err != nil {
		return err
	}
	return nil
}
