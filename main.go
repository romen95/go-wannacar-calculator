package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-wannacar-calculator/bot"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Функция для получения данных из Google Sheets
func getSheetData() ([]map[string]string, error) {
	ctx := context.Background()

	// Укажите путь к вашему файлу с учетными данными
	credsFile := "internal/credentials.json"

	// Создаем клиент для работы с Google Sheets API
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать клиент: %v", err)
	}

	// Укажите ID вашей таблицы и диапазон (например, "Лист1!A1:C10")
	spreadsheetId := "1QsOurjcx5quV-zP-KqQCDpGroNvg5btWvsYfbXEwvIc"
	readRange := "Заказы!A1:C10"

	// Получаем данные из таблицы
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить данные: %v", err)
	}

	// Преобразуем данные в удобный формат
	var orders []map[string]string
	if len(resp.Values) > 0 {
		headers := resp.Values[0] // Первая строка — заголовки
		for _, row := range resp.Values[1:] {
			order := make(map[string]string)
			for i, value := range row {
				order[headers[i].(string)] = value.(string)
			}
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// Обработчик для API, возвращающего данные о заказах
func ordersHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем данные из Google Sheets
	orders, err := getSheetData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
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

	// API для получения данных о заказах
	http.HandleFunc("/orders", ordersHandler)

	log.Println("Сервер запущен на порту 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
