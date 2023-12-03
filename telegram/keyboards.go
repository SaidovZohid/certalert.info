package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var chooseLanguageStart = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(DisplayLangUz),
		tgbotapi.NewKeyboardButton(DisplayLangRu),
		tgbotapi.NewKeyboardButton(DisplayLangEng),
	),
)
