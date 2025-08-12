package bot

import (
	"fmt"
	"go-wannacar-calculator/database"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	Bot *tgbotapi.BotAPI
	DB  *database.DB
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
	chatID := message.Chat.ID
	user := h.DB.GetUserByID(chatID)
	if user == nil {
		err := h.DB.CreateUser(chatID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при создании пользователя.")
			if _, err := h.Bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}
			return
		}
	}
	text := "Привет, я твой личный помощник в покупке автомобиля!\nНажми на кнопку ниже, чтобы открыть приложение👇🏻"

	miniAppURL := "t.me/wanna_car_bot/app"

	// Создаем inline-кнопку
	btn := tgbotapi.NewInlineKeyboardButtonURL("Открыть мини-приложение", miniAppURL)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))

	// Создаем сообщение с кнопкой
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if _, err := h.Bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func (h *BotHandler) SendResponse(chatID int64, response map[string]float64, country, typeAuto string, priceWon, priceEuro, engineVolume float64, yearOfManufacture int) {
	user := h.DB.GetUserByID(chatID)
	if user == nil {
		err := h.DB.CreateUser(chatID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при создании пользователя.")
			if _, err := h.Bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}
			return
		}
	}

	// Проверяем, есть ли данные в response
	if len(response) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Нет данных для отображения.")
		if _, err := h.Bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		return
	}

	// Мапа с русскими названиями для каждого ключа
	fieldNames := map[string]string{
		"koreaCustomsСostPrice":         "Таможенная стоимость",
		"germanyCustomsСostPrice":       "Таможенная стоимость",
		"exchangeRateWonToRub":          "Курс вон к рублю (1000 ₩)",
		"exchangeRateEuroToRub":         "Курс евро к рублю",
		"koreaRecyclingCollection":      "Утильсбор",
		"germanyRecyclingCollection":    "Утильсбор",
		"koreaPriceAutoRub":             "Стоимость авто в Корее",
		"germanyPriceAutoRub":           "Стоимость авто в Германии",
		"koreaLogisticPriceRub":         "Логистика до Владивостока",
		"germanyLogisticMoscowPriceRub": "Логистика до СПБ",
		"koreaLogisticsDestination":     "Логистика до места назначенния",
		"germanyLogisticsDestination":   "Логистика до места назначенния",
		"koreaCommission":               "Комиссия",
		"germanyCommission":             "Комиссия",
		"koreaDocumentsPrice":           "Оформление документов",
		"germanyDocumentsPrice":         "Оформление документов",
		"koreaResultPrice":              "Итоговая стоимость",
		"germanyResultPrice":            "Итоговая стоимость",
		"germanyPricePercentRub":        "Услуги брокера в Германии",
	}

	// Срез с ключами в нужном порядке
	orderedKeys := []string{
		"koreaResultPrice",
		"germanyResultPrice",
		"koreaPriceAutoRub",
		"germanyPriceAutoRub",
		"germanyPricePercentRub",
		"koreaCustomsСostPrice",
		"germanyCustomsСostPrice",
		"koreaRecyclingCollection",
		"germanyRecyclingCollection",
		"koreaLogisticPriceRub",
		"germanyLogisticMoscowPriceRub",
		"koreaLogisticsDestination",
		"germanyLogisticsDestination",
		"koreaCommission",
		"germanyCommission",
		"koreaDocumentsPrice",
		"germanyDocumentsPrice",
		"exchangeRateWonToRub",
		"exchangeRateEuroToRub",
	}

	// Формируем текстовое сообщение
	var messageText string

	calculationCount := user.CalculateCount

	messageText += fmt.Sprintf("Расчет стоимости №%d", calculationCount)
	log.Printf("Расчет стоимости №%d", calculationCount)

	// Добавляем разделитель
	messageText += "\n\n"

	// Формируем текстовое сообщение из response
	for _, key := range orderedKeys {
		if value, exists := response[key]; exists {
			if russianName, ok := fieldNames[key]; ok {
				// Проверяем, является ли ключ курсом валюты
				if key == "exchangeRateWonToRub" || key == "exchangeRateEuroToRub" {
					messageText += "\n"
					// Для курсов валют
					messageText += fmt.Sprintf("%s:\n%.2f ₽\n", russianName, value)
				} else if key == "koreaResultPrice" || key == "germanyResultPrice" {
					messageText += fmt.Sprintf("%s:\n%.0f ₽\n\nВ итоговую стоимость входит:\n", russianName, value)
				} else {
					// Для остальных значений
					messageText += fmt.Sprintf("%s:\n%.0f ₽\n", russianName, value)
				}
			} else {
				// Если ключ не найден в мапе, используем оригинальный ключ
				messageText += fmt.Sprintf("%s:\n%.0f ₽\n", key, value)
			}
		}
	}

	// Добавляем разделитель
	messageText += "\n"

	// Добавляем информацию о стране, типе авто и годе выпуска
	messageText += fmt.Sprintf("Страна: %s\n", country)
	messageText += fmt.Sprintf("Тип авто: %s\n", typeAuto)
	messageText += fmt.Sprintf("Год выпуска: %d г.\n", yearOfManufacture)

	// Добавляем информацию об объеме двигателя, если он есть
	if engineVolume > 0 {
		messageText += fmt.Sprintf("Объем двигателя: %.0f см³\n", engineVolume)
	}

	// Добавляем информацию о цене в вонах (только для Кореи)
	if country == "Корея" && priceWon > 0 {
		messageText += fmt.Sprintf("Цена в вонах: %.0f ₩\n", priceWon)
	}

	// Добавляем информацию о цене в евро (только для Германии)
	if country == "Германия" && priceEuro > 0 {
		messageText += fmt.Sprintf("Цена в евро: %.0f €\n", priceEuro)
	}

	// Отправляем сообщение
	msg := tgbotapi.NewMessage(chatID, messageText)
	if _, err := h.Bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}

	// Обновляем счетчик расчетов
	h.DB.UpdateCalculateCount(chatID)
}
