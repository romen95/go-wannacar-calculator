package internal

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/api/sheets/v4"
)

// Конфигурация Google Sheets
const (
	SpreadsheetID   = "1WpF2WzGZwCQgnLUsB8d5_GnMbeqUE5TIxP6fO7QNbu4" // ID вашей таблицы
	CredentialsFile = "internal/credentials.json"                    // Путь к файлу с учетными данными
)

// Структура для хранения данных о таможенных ставках
type CustomsRate struct {
	Min  float64
	Max  float64
	Rate float64
}

// Функция для получения значения из Google Sheets
func GetValueFromSheet(srv *sheets.Service, sheetName, cellRange string) (float64, error) {
	readRange := fmt.Sprintf("%s!%s", sheetName, cellRange)
	resp, err := srv.Spreadsheets.Values.Get(SpreadsheetID, readRange).Do()
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

// Функция для получения утильсбора авто до 3х лет из Google Sheets
func GetRecyclingUpTo3YearsRates(srv *sheets.Service) ([]CustomsRate, error) {
	readRange := "Утильсбор!B3:D7"
	resp, err := srv.Spreadsheets.Values.Get(SpreadsheetID, readRange).Do()
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

// Функция для получения утильсбора авто до 3х лет из Google Sheets
func GetRecyclingFrom3YearsRates(srv *sheets.Service) ([]CustomsRate, error) {
	readRange := "Утильсбор!B10:D14"
	resp, err := srv.Spreadsheets.Values.Get(SpreadsheetID, readRange).Do()
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

// Функция для получения таможенных ставок авто до 3х лет из Google Sheets
func GetCustomsUpTo3YearsRates(srv *sheets.Service) ([]CustomsRate, error) {
	readRange := "Таможня!B3:D8"
	resp, err := srv.Spreadsheets.Values.Get(SpreadsheetID, readRange).Do()
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

// Функция для получения таможенных ставок авто от 3х до 5ти лет из Google Sheets
func GetCustomsFrom3To5YearsRates(srv *sheets.Service) ([]CustomsRate, error) {
	readRange := "Таможня!B11:D16"
	resp, err := srv.Spreadsheets.Values.Get(SpreadsheetID, readRange).Do()
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

// Функция для получения таможенных ставок авто от 5ти лет из Google Sheets
func GetCustomsFrom5YearsRates(srv *sheets.Service) ([]CustomsRate, error) {
	readRange := "Таможня!B19:D24"
	resp, err := srv.Spreadsheets.Values.Get(SpreadsheetID, readRange).Do()
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
