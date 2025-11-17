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
func calculateGeorgiaRecyclingCollection(typeAuto string, carAge int, engineVolume float64, srv *sheets.Service) (float64, error) {
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
func calculateGeorgiaCustomsСostPrice(typeAuto string, priceDollar float64, carAge int, engineVolume float64, srv *sheets.Service) (float64, float64, float64, error) {
	exchangeRateDollarToRub, err := internal.GetValueFromSheet(srv, "Курс", "B2")
	if err != nil {
		return 0, 0, 0, err
	}
	log.Printf("$: %f ₽\n", exchangeRateDollarToRub)

	// Получаем значение из ячейки "Курс!B3" (курс евро к рублю)
	exchangeRateEuroToRub, err := internal.GetValueFromSheet(srv, "Курс", "B3")
	if err != nil {
		return 0, 0, 0, err
	}
	log.Printf("€: %f ₽\n", exchangeRateEuroToRub)

	log.Printf("Стоимость авто в долларах: %.2f $\n", priceDollar)
	log.Printf("Возраст авто: %d\n", carAge)
	log.Printf("Объем двигателя в см³: %.2f см³\n", engineVolume)

	// Преобразуем стоимость авто в евро
	priceEuro := (priceDollar * exchangeRateDollarToRub) / exchangeRateEuroToRub
	priceEuroRounded := math.Ceil(priceEuro) // Округляем в большую сторону

	log.Printf("Стоимость авто в евро (округлено в большую сторону): %.2f €\n", priceEuroRounded)

	var customsСostPrice float64

	switch typeAuto {
	case "Бензиновый/дизельный двигатель":
		// Если возраст авто меньше 3 лет
		if carAge < 3 {
			customsRates, err := internal.GetCustomsUpTo3YearsRates(srv)
			if err != nil {
				return 0, 0, 0, err
			}

			var rate float64
			for _, row := range customsRates {
				if priceEuroRounded >= row.Min && priceEuroRounded <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, 0, 0, fmt.Errorf("не удалось найти подходящую ставку для стоимости %.2f евро", priceEuroRounded)
			}

			customsСostPrice = priceDollar * exchangeRateDollarToRub * (rate / 100)
			log.Printf("Таможенная ставка рассчитанная от стоимости авто: %.2f ₽\n", customsСostPrice)
		}

		// Если возраст авто меньше от 3х до 5ти
		if carAge >= 3 && carAge <= 5 {
			customsRates, err := internal.GetCustomsFrom3To5YearsRates(srv)
			if err != nil {
				return 0, 0, 0, err
			}

			var rate float64
			for _, row := range customsRates {
				if engineVolume >= row.Min && engineVolume <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, 0, 0, fmt.Errorf("не удалось найти подходящую ставку для стоимости %.2f евро", priceEuroRounded)
			}

			customsСostPrice = engineVolume * rate * exchangeRateEuroToRub
			log.Printf("Таможенная ставка рассчитанная по см³: %.2f\n", customsСostPrice)
		}

		// Если возраст авто меньше от 5ти
		if carAge > 5 {
			customsRates, err := internal.GetCustomsFrom5YearsRates(srv)
			if err != nil {
				return 0, 0, 0, err
			}

			var rate float64
			for _, row := range customsRates {
				if engineVolume >= row.Min && engineVolume <= row.Max {
					rate = row.Rate
					break
				}
			}

			if rate == 0 {
				return 0, 0, 0, fmt.Errorf("не удалось найти подходящую ставку для стоимости %.2f евро", priceEuroRounded)
			}

			customsСostPrice = engineVolume * rate * exchangeRateEuroToRub
			log.Printf("Таможенная ставка рассчитанная по см³: %.2f\n", customsСostPrice)
		}
	case "Электромобиль":
		rate, err := internal.GetValueFromSheet(srv, "Таможня", "G3")
		if err != nil {
			return 0, 0, 0, err
		}

		customsСostPrice = priceDollar * exchangeRateDollarToRub * (rate / 100)
		log.Printf("Таможенная ставка рассчитанная от стоимости электромобиля: %.2f ₽\n", customsСostPrice)

	}

	return customsСostPrice, exchangeRateDollarToRub, exchangeRateEuroToRub, nil
}

func CalculateGeorgiaResult(typeAuto string, priceDollar float64, yearOfManufacture int, engineVolume float64, logisticsDestination float64, srv *sheets.Service) (map[string]float64, error) {
	// Инициализируем map для хранения результатов
	result := make(map[string]float64)

	// Получаем текущий год
	currentYear := time.Now().Year()

	// Рассчитываем возраст авто
	carAge := currentYear - yearOfManufacture

	// Рассчитываем стоимость таможенной ставки
	georgiaCustomsСostPrice, exchangeRateDollarToRub, exchangeRateEuroToRub, err := calculateGeorgiaCustomsСostPrice(typeAuto, priceDollar, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости таможенной ставки грузинского авто: %v\n", err)
		return nil, err
	}
	result["georgiaCustomsСostPrice"] = RoundUpTo10000(georgiaCustomsСostPrice)

	result["exchangeRateDollarToRub"] = math.Round(exchangeRateDollarToRub*100) / 100
	result["exchangeRateEuroToRub"] = math.Round(exchangeRateEuroToRub*100) / 100

	// Рассчитываем стоимость утильсбора
	georgiaRecyclingCollection, err := calculateGeorgiaRecyclingCollection(typeAuto, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости утильсбора грузинского авто: %v\n", err)
		return nil, err
	}
	result["georgiaRecyclingCollection"] = RoundUpTo10000(georgiaRecyclingCollection)

	// Получаем стоимость логистики из Грузии
	logisticGeorgiaPrice, err := internal.GetValueFromSheet(srv, "Грузия", "B2")
	if err != nil {
		return nil, err
	}

	georgiaPriceAutoRub := priceDollar * exchangeRateDollarToRub
	result["georgiaPriceAutoRub"] = RoundUpTo10000(georgiaPriceAutoRub)

	// Рассчитываем стоимость логистики из Грузии в рублях
	georgiaLogisticPriceRub := logisticGeorgiaPrice * exchangeRateDollarToRub
	result["georgiaLogisticPriceRub"] = RoundUpTo10000(georgiaLogisticPriceRub)

	result["georgiaLogisticsDestination"] = math.Ceil(logisticsDestination)

	// Получаем комиссию
	georgiaCommission, err := internal.GetValueFromSheet(srv, "Грузия", "C3")
	if err != nil {
		return nil, err
	}
	result["georgiaCommission"] = georgiaCommission

	// Получаем стоимость оформления документов
	georgiaDocumentsPrice, err := internal.GetValueFromSheet(srv, "Грузия", "E3")
	if err != nil {
		return nil, err
	}
	result["georgiaDocumentsPrice"] = georgiaDocumentsPrice

	// Рассчитываем итоговую стоимость
	georgiaResultPrice := RoundUpTo10000(georgiaCustomsСostPrice) + RoundUpTo10000(georgiaRecyclingCollection) + RoundUpTo10000(georgiaPriceAutoRub) + RoundUpTo10000(georgiaLogisticPriceRub) + georgiaCommission + math.Ceil(logisticsDestination) + georgiaDocumentsPrice
	result["georgiaResultPrice"] = georgiaResultPrice

	// Возвращаем map с результатами
	return result, nil
}
