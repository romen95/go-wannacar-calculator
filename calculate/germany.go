package calculate

import (
	"fmt"
	"log"
	"math"
	"time"

	"go-wannacar-calculator/internal"

	"google.golang.org/api/sheets/v4"
)

// Функция для расчета утильсбора
func calculateGermanyRecyclingCollection(typeAuto string, carAge int, engineVolume float64, srv *sheets.Service) (float64, error) {
	var recyclingCollectionPrice float64

	switch typeAuto {
	case "Бензиновый/дизельный двигатель":
		// Если возраст авто меньше 3 лет
		if carAge < 3 {
			recyclingRates, err := internal.GetRecyclingUpTo3YearsRates(srv)
			if err != nil {
				return 0, err
			}

			var rate float64
			for _, row := range recyclingRates {
				if engineVolume >= row.Min && engineVolume <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, fmt.Errorf("не удалось найти подходящую ставку для объема %.2f см³", engineVolume)
			}

			log.Printf("Найден утильсбор: %.2f ₽\n", rate)

			recyclingCollectionPrice = rate
			log.Printf("Утильсбор рассчитан от объема: %.2f ₽\n", recyclingCollectionPrice)
		} else {
			// Если возраст авто больше или равен 3 годам
			recyclingRates, err := internal.GetRecyclingFrom3YearsRates(srv)
			if err != nil {
				return 0, err
			}

			var rate float64
			for _, row := range recyclingRates {
				if engineVolume >= row.Min && engineVolume <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, fmt.Errorf("не удалось найти подходящую ставку для объема %.2f см³", engineVolume)
			}

			log.Printf("Найден утильсбор: %.2f ₽\n", rate)

			recyclingCollectionPrice = rate
			log.Printf("Утильсбор рассчитан от объема: %.2f ₽\n", recyclingCollectionPrice)
		}
	case "Электромобиль":
		if carAge < 3 {
			rate, err := internal.GetValueFromSheet(srv, "Утильсбор", "G3")
			if err != nil {
				return 0, err
			}

			recyclingCollectionPrice = rate
			log.Printf("Утильсбор рассчитан от объема: %.2f ₽\n", recyclingCollectionPrice)
		} else {
			rate, err := internal.GetValueFromSheet(srv, "Утильсбор", "G6")
			if err != nil {
				return 0, err
			}

			recyclingCollectionPrice = rate
			log.Printf("Утильсбор рассчитан от объема: %.2f ₽\n", recyclingCollectionPrice)
		}
	}

	return recyclingCollectionPrice, nil
}

// Функция для расчета таможенной стоимости
func calculateGermanyCustomsСostPrice(typeAuto string, priceEuro float64, carAge int, engineVolume float64, srv *sheets.Service) (float64, float64, error) {
	// Получаем значение из ячейки "Курс!B3" (курс евро к рублю)
	exchangeRateEuroToRub, err := internal.GetValueFromSheet(srv, "Курс", "B3")
	if err != nil {
		return 0, 0, err
	}
	log.Printf("€: %f ₽\n", exchangeRateEuroToRub)

	log.Printf("Стоимость авто в евро: %.2f €\n", priceEuro)
	log.Printf("Возраст авто: %d\n", carAge)
	log.Printf("Объем двигателя в см³: %.2f см³\n", engineVolume)

	var customsСostPrice float64

	switch typeAuto {
	case "Бензиновый/дизельный двигатель":
		// Если возраст авто меньше 3 лет
		if carAge < 3 {
			customsRates, err := internal.GetCustomsUpTo3YearsRates(srv)
			if err != nil {
				return 0, 0, err
			}

			var rate float64
			for _, row := range customsRates {
				if priceEuro >= row.Min && priceEuro <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, 0, fmt.Errorf("не удалось найти подходящую ставку для стоимости %.2f евро", priceEuro)
			}

			log.Printf("Найдена таможенная ставка: %.2f%%\n", rate)

			customsСostPrice = priceEuro * exchangeRateEuroToRub * (rate / 100)
			log.Printf("Таможенная ставка рассчитанная от стоимости авто: %.2f ₽\n", customsСostPrice)
		}

		// Если возраст авто меньше от 3х до 5ти
		if carAge >= 3 && carAge <= 5 {
			customsRates, err := internal.GetCustomsFrom3To5YearsRates(srv)
			if err != nil {
				return 0, 0, err
			}

			var rate float64
			for _, row := range customsRates {
				if engineVolume >= row.Min && engineVolume <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, 0, fmt.Errorf("не удалось найти подходящую ставку для стоимости %.2f евро", priceEuro)
			}

			log.Printf("Найдена таможенная ставка: %.2f € за 1 см³\n", rate)

			customsСostPrice = engineVolume * rate * exchangeRateEuroToRub
			log.Printf("Таможенная ставка рассчитанная по см³: %.2f\n", customsСostPrice)
		}

		// Если возраст авто меньше от 5ти
		if carAge > 5 {
			customsRates, err := internal.GetCustomsFrom5YearsRates(srv)
			if err != nil {
				return 0, 0, err
			}

			var rate float64
			for _, row := range customsRates {
				if engineVolume >= row.Min && engineVolume <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, 0, fmt.Errorf("не удалось найти подходящую ставку для стоимости %.2f евро", priceEuro)
			}

			log.Printf("Найдена таможенная ставка: %.2f € за 1 см³\n", rate)

			customsСostPrice = engineVolume * rate * exchangeRateEuroToRub
			log.Printf("Таможенная ставка рассчитанная по см³: %.2f\n", customsСostPrice)
		}
	case "Электромобиль":
		rate, err := internal.GetValueFromSheet(srv, "Таможня", "G3")
		if err != nil {
			return 0, 0, err
		}

		customsСostPrice = priceEuro * exchangeRateEuroToRub * (rate / 100)
		log.Printf("Таможенная ставка рассчитанная от стоимости электромобиля: %.2f ₽\n", customsСostPrice)

	}

	return customsСostPrice, exchangeRateEuroToRub, nil
}

func CalculateGermanyResult(typeAuto string, priceEuro float64, yearOfManufacture int, engineVolume float64, srv *sheets.Service) (map[string]float64, error) {
	// Инициализируем map для хранения результатов
	result := make(map[string]float64)

	// Получаем текущий год
	currentYear := time.Now().Year()

	// Рассчитываем возраст авто
	carAge := currentYear - yearOfManufacture

	// Рассчитываем стоимость таможенной ставки
	germanyCustomsСostPrice, exchangeRateEuroToRub, err := calculateGermanyCustomsСostPrice(typeAuto, priceEuro, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости таможенной ставки немецкого авто: %v\n", err)
		return nil, err
	}
	result["germanyCustomsСostPrice"] = math.Ceil(germanyCustomsСostPrice)

	result["exchangeRateEuroToRub"] = exchangeRateEuroToRub

	// Рассчитываем стоимость утильсбора
	germanyRecyclingCollection, err := calculateGermanyRecyclingCollection(typeAuto, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости утильсбора немецкого авто: %v\n", err)
		return nil, err
	}
	result["germanyRecyclingCollection"] = math.Ceil(germanyRecyclingCollection)

	// Получаем процент наценки
	percent, err := internal.GetValueFromSheet(srv, "Германия", "F2")
	if err != nil {
		return nil, err
	}

	pricePercentRub := priceEuro * (percent / 100) * exchangeRateEuroToRub
	result["pricePercentRub"] = math.Ceil(pricePercentRub)

	priceAutoRub := priceEuro * exchangeRateEuroToRub
	result["priceAutoRub"] = math.Ceil(priceAutoRub)

	// Получаем комиссию
	commission, err := internal.GetValueFromSheet(srv, "Германия", "C3")
	if err != nil {
		return nil, err
	}
	result["commission"] = commission

	// Получаем стоимость логистики по Москве
	logisticMoscowPrice, err := internal.GetValueFromSheet(srv, "Германия", "B2")
	if err != nil {
		return nil, err
	}

	logisticMoscowPriceRub := logisticMoscowPrice * exchangeRateEuroToRub
	result["logisticMoscowPriceRub"] = math.Ceil(logisticMoscowPriceRub)

	// Получаем стоимость оформления документов
	documentsPrice, err := internal.GetValueFromSheet(srv, "Германия", "D3")
	if err != nil {
		return nil, err
	}
	result["documentsPrice"] = documentsPrice

	// Рассчитываем итоговую стоимость
	resultPrice := math.Ceil(priceAutoRub) + math.Ceil(pricePercentRub) + math.Ceil(germanyCustomsСostPrice) + math.Ceil(germanyRecyclingCollection) + math.Ceil(logisticMoscowPriceRub) + commission + documentsPrice
	result["resultPrice"] = resultPrice

	// Возвращаем map с результатами
	return result, nil
}
