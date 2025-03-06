package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-wannacar-calculator/bot"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Конфигурация Google Sheets
const (
	spreadsheetID   = "1WpF2WzGZwCQgnLUsB8d5_GnMbeqUE5TIxP6fO7QNbu4" // ID вашей таблицы
	credentialsFile = "internal/credentials.json"                    // Путь к файлу с учетными данными
)

// Функция для получения значения из Google Sheets
func getValueFromSheet(srv *sheets.Service, sheetName, cellRange string) (float64, error) {
	readRange := fmt.Sprintf("%s!%s", sheetName, cellRange)
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("не удалось получить данные: %v", err)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return 0, fmt.Errorf("ячейка %s пуста", readRange)
	}

	value := resp.Values[0][0].(string)
	// Заменяем запятые на точки
	value = strings.Replace(value, ",", ".", -1)

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("не удалось преобразовать значение в число: %v", err)
	}

	return floatValue, nil
}

// Функция для расчета стоимости авто для Кореи
func calculateKoreaPrice(priceWon float64, carAge int, srv *sheets.Service) (float64, error) {
	// Получаем значение из ячейки "Курс!B4" (курс вон к рублю)
	exchangeRateWonToRub, err := getValueFromSheet(srv, "Курс", "B4")
	if err != nil {
		return 0, err
	}

	// Получаем значение из ячейки "Курс!B3" (курс евро к рублю)
	exchangeRateEuroToRub, err := getValueFromSheet(srv, "Курс", "B3")
	if err != nil {
		return 0, err
	}

	// Преобразуем стоимость авто в евро
	priceEuro := (priceWon * exchangeRateWonToRub) / exchangeRateEuroToRub
	priceEuroRounded := math.Ceil(priceEuro) // Округляем в большую сторону

	log.Printf("Стоимость авто в евро (округлено): %.2f\n", priceEuroRounded)

	// Если возраст авто меньше 3 лет, применяем таможенную ставку
	if carAge < 3 {
		// Получаем данные из таблицы "Таможня"
		customsRates, err := getCustomsRates(srv)
		if err != nil {
			return 0, err
		}

		// Находим подходящую ставку
		var rate float64
		for _, row := range customsRates {
			if priceEuroRounded >= row.Min && priceEuroRounded <= row.Max {
				rate = row.Rate
				break
			}
		}

		if rate == 0 {
			return 0, fmt.Errorf("не удалось найти подходящую ставку для стоимости %.2f евро", priceEuroRounded)
		}

		log.Printf("Найдена таможенная ставка: %.2f%%\n", rate)

		// Получаем значение из ячейки "Корея!B2"
		koreaCost, err := getValueFromSheet(srv, "Корея", "B2")
		if err != nil {
			return 0, err
		}

		// Рассчитываем итоговую стоимость
		totalPrice := (priceWon + koreaCost) * exchangeRateWonToRub * (rate / 100)
		return totalPrice, nil
	}

	// Если возраст авто больше или равен 3 годам, возвращаем базовую стоимость
	return priceWon * exchangeRateWonToRub, nil
}

// Структура для хранения данных о таможенных ставках
type CustomsRate struct {
	Min  float64
	Max  float64
	Rate float64
}

// Функция для получения таможенных ставок из Google Sheets
func getCustomsRates(srv *sheets.Service) ([]CustomsRate, error) {
	readRange := "Таможня!B3:D8"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить данные: %v", err)
	}

	var rates []CustomsRate
	for _, row := range resp.Values {
		min, err := strconv.ParseFloat(strings.Replace(row[0].(string), ",", ".", -1), 64)
		if err != nil {
			return nil, fmt.Errorf("ошибка преобразования Min: %v", err)
		}

		max, err := strconv.ParseFloat(strings.Replace(row[1].(string), ",", ".", -1), 64)
		if err != nil {
			return nil, fmt.Errorf("ошибка преобразования Max: %v", err)
		}

		rateStr := strings.Replace(row[2].(string), "%", "", -1)
		rate, err := strconv.ParseFloat(strings.Replace(rateStr, ",", ".", -1), 64)
		if err != nil {
			return nil, fmt.Errorf("ошибка преобразования Rate: %v", err)
		}

		rates = append(rates, CustomsRate{Min: min, Max: max, Rate: rate})
	}

	return rates, nil
}

// Обработчик для API расчета стоимости
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS (если запрос из браузера)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обрабатываем OPTIONS-запрос для CORS
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Парсим входные данные
	var requestData struct {
		PriceWon float64 `json:"priceWon"`
		CarAge   int     `json:"carAge"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Printf("Ошибка при декодировании JSON: %v\n", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Создаем клиент для работы с Google Sheets API
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Printf("Ошибка при создании клиента Google Sheets: %v\n", err)
		http.Error(w, fmt.Sprintf("Не удалось создать клиент: %v", err), http.StatusInternalServerError)
		return
	}

	// Выполняем расчет стоимости
	totalPrice, err := calculateKoreaPrice(requestData.PriceWon, requestData.CarAge, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости: %v\n", err)
		http.Error(w, fmt.Sprintf("Ошибка расчета: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем результат в формате JSON
	response := map[string]float64{"totalPrice": totalPrice}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Рассчитанная стоимость: %.2f\n", totalPrice)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки файла .env")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	handler := bot.NewBotHandler(botToken)

	go func() {
		log.Println("Запуск Telegram-бота...")
		handler.Run()
	}()

	// Отдача статических файлов из папки "img"
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))

	// Обработка HTML-шаблона
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			log.Printf("Ошибка при загрузке шаблона: %v", err)
			http.Error(w, "Ошибка при загрузке страницы", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			log.Printf("Ошибка при выполнении шаблона: %v", err)
			http.Error(w, "Ошибка при отображении страницы", http.StatusInternalServerError)
		}
	})

	// API для расчета стоимости
	http.HandleFunc("/calculate", calculateHandler)

	// Запуск сервера
	log.Println("Сервер запущен на порту 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
