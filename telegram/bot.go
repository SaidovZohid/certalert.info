package telegram

import (
	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot       *tgbotapi.BotAPI
	appLogger *logger.Logger
	cfg       *config.Config
}

func NewBot(bot *tgbotapi.BotAPI, appLogger *logger.Logger, cfg *config.Config) *Bot {
	return &Bot{
		bot:       bot,
		appLogger: appLogger,
		cfg:       cfg,
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
		if update.Message.Command() == "start" {
			// Extract the start parameter value if it exists
			startParam := update.Message.CommandArguments()
			
			

			// Access the start parameter for further processing
			b.appLogger.Infof("Received start parameter: %s", startParam)

			// Perform actions based on the start parameter
			// ...
		} 
	}
}

func (b *Bot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.bot.GetUpdatesChan(u)
}

// func (b *Bot) handleSendTextMessage(chatID int64, message string) {
// 	msg := tgbotapi.NewMessage(chatID, message)

// 	b.bot.Send(msg)
// }
