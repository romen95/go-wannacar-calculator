package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"go-wannacar-calculator/bot"
	"go-wannacar-calculator/calculate"
	"go-wannacar-calculator/database"
	"go-wannacar-calculator/internal"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Обработчик для API расчета стоимости
func calculateHandler(h *bot.BotHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем CORS (если запрос из браузера)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обрабатываем OPTIONS-запрос для CORS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var requestData struct {
			Country              string  `json:"country"`
			TypeAuto             string  `json:"typeAuto"`
			PriceWon             float64 `json:"priceWon"`
			PriceEuro            float64 `json:"priceEuro"`
			YearOfManufacture    int     `json:"yearOfManufacture"`
			EngineVolume         float64 `json:"engineVolume"`
			LogisticsDestination float64 `json:"logisticsDestination"`
			ChatID               int64   `json:"chatID"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			log.Printf("Ошибка при декодировании JSON: %v\n", err)
			http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
			return
		}

		// Создаем клиент для работы с Google Sheets API
		ctx := context.Background()
		srv, err := sheets.NewService(ctx, option.WithCredentialsFile(internal.CredentialsFile))
		if err != nil {
			log.Printf("Ошибка при создании клиента Google Sheets: %v\n", err)
			http.Error(w, fmt.Sprintf("Не удалось создать клиент: %v", err), http.StatusInternalServerError)
			return
		}

		var response map[string]float64

		switch requestData.Country {
		case "Корея":
			response, err = calculate.CalculateKoreaResult(requestData.TypeAuto, requestData.PriceWon, requestData.YearOfManufacture, requestData.EngineVolume, requestData.LogisticsDestination, srv)
			if err != nil {
				log.Printf("Ошибка при передаче данных на фронтенд: %v\n", err)
				http.Error(w, fmt.Sprintf("Не удалось передать данные: %v", err), http.StatusInternalServerError)
				return
			}
		case "Германия":
			response, err = calculate.CalculateGermanyResult(requestData.TypeAuto, requestData.PriceEuro, requestData.YearOfManufacture, requestData.EngineVolume, requestData.LogisticsDestination, srv)
			if err != nil {
				log.Printf("Ошибка при передаче данных на фронтенд: %v\n", err)
				http.Error(w, fmt.Sprintf("Не удалось передать данные: %v", err), http.StatusInternalServerError)
				return
			}
		}
		go h.SendResponse(requestData.ChatID, response, requestData.Country, requestData.TypeAuto, requestData.PriceWon, requestData.PriceEuro, requestData.EngineVolume, requestData.YearOfManufacture)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func main() {
	// Подключение к базе данных
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки файла .env")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	handler := bot.NewBotHandler(db, botToken)

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
	http.HandleFunc("/calculate", calculateHandler(handler))

	// Запуск сервера
	log.Println("Сервер запущен на порту 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
