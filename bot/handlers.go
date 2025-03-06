package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	Bot *tgbotapi.BotAPI
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		h.HandleMessage(update.Message)
	}
}

func (h *BotHandler) HandleMessage(message *tgbotapi.Message) {
	switch message.Text {
	case "/start":
		h.HandleStart(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Введите /start")
		if _, err := h.Bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
	}
}

func (h *BotHandler) HandleStart(message *tgbotapi.Message) {
	text := "Здарова"

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	if _, err := h.Bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
