package telegram

import (
	"context"
	"errors"
	"strconv"

	"github.com/SaidovZohid/certalert.info/storage/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v4"
)

func (b *Bot) handleCommandsStart(update tgbotapi.Update) bool {
	tgUser, err := b.strg.Integrations().GetFromTelegramByTGID(context.Background(), update.Message.From.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		b.appLogger.Errorf("Failed to get from telegram users %s", err.Error())
		b.handleSendTextMessage(update.Message.Chat.ID, uzInternalErrorMsg+"\n\n"+engInternalErrorMsg+"\n\n"+ruInternalErrorMsg)
		return true
	}

	if tgUser != nil {
		b.handleSendTextMessage(update.Message.Chat.ID, uzAlreadyLinkedToThisAccount+"\n\n"+engAlreadyLinkedToThisAccount+"\n\n"+ruAlreadyLinkedToThisAccount)
		return true
	}

	// Extract the start parameter value if it exists
	startParam := update.Message.CommandArguments()

	userId, err := strconv.Atoi(startParam)
	if err != nil {
		b.handleSendTextMessage(update.Message.Chat.ID, uzNotValidUserID+"\n\n"+engNotValidUserID+"\n\n"+ruNotValidUserID)
		return true
	}

	user, err := b.strg.User().GetUserByID(context.Background(), int64(userId))
	if err != nil {
		b.appLogger.Errorf("Failed to get user %s", err.Error())
		b.handleSendTextMessage(update.Message.Chat.ID, uzNotFoundUserID+"\n\n"+engNotFoundUserID+"\n\n"+ruNotFoundUserID)
		return true
	}

	err = b.strg.Integrations().LinkTelegramAccountToWebsiteAccount(context.Background(), &models.TelegramUser{
		UserID:         user.ID,
		ChatID:         update.Message.Chat.ID,
		TelegramUserID: update.Message.From.ID,
		Lang:           langEng,
		Step:           stepLang,
	})
	if err != nil {
		b.appLogger.Errorf("Failed to create integration for telegram and user %s", err.Error())
		return true
	}
	if err = b.strg.Notifications().UpdateTheAlertIntegrations(context.Background(), user.ID, telegramAlert, true); err != nil {
		b.appLogger.Errorf("Failed to create integration for telegram and user %s", err.Error())
		return true
	}

	if err := b.handleSuccessfullyLinkedAccount(&update); err != nil {
		b.appLogger.Errorf("Failed to send message about successfully linked account %s", err.Error())
	}
	return true
}
