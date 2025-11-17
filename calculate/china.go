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
func calculateChinaRecyclingCollection(typeAuto string, carAge int, engineVolume float64, srv *sheets.Service) (float64, error) {
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
func calculateChinaCustomsСostPrice(typeAuto string, priceYuan float64, carAge int, engineVolume float64, srv *sheets.Service) (float64, float64, float64, error) {
	exchangeRateYuanToRub, err := internal.GetValueFromSheet(srv, "Курс", "B5")
	if err != nil {
		return 0, 0, 0, err
	}
	log.Printf("¥: %f ₽\n", exchangeRateYuanToRub)

	// Получаем значение из ячейки "Курс!B3" (курс евро к рублю)
	exchangeRateEuroToRub, err := internal.GetValueFromSheet(srv, "Курс", "B3")
	if err != nil {
		return 0, 0, 0, err
	}
	log.Printf("€: %f ₽\n", exchangeRateEuroToRub)

	log.Printf("Стоимость авто в юанях: %.2f ¥\n", priceYuan)
	log.Printf("Возраст авто: %d\n", carAge)
	log.Printf("Объем двигателя в см³: %.2f см³\n", engineVolume)

	// Преобразуем стоимость авто в евро
	priceEuro := (priceYuan * exchangeRateYuanToRub) / exchangeRateEuroToRub
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

			customsСostPrice = priceYuan * exchangeRateYuanToRub * (rate / 100)
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

		customsСostPrice = priceYuan * exchangeRateYuanToRub * (rate / 100)
		log.Printf("Таможенная ставка рассчитанная от стоимости электромобиля: %.2f ₽\n", customsСostPrice)

	}

	return customsСostPrice, exchangeRateYuanToRub, exchangeRateEuroToRub, nil
}

func CalculateChinaResult(typeAuto string, priceYuan float64, yearOfManufacture int, engineVolume float64, logisticsDestination float64, srv *sheets.Service) (map[string]float64, error) {
	// Инициализируем map для хранения результатов
	result := make(map[string]float64)

	// Получаем текущий год
	currentYear := time.Now().Year()

	// Рассчитываем возраст авто
	carAge := currentYear - yearOfManufacture

	// Рассчитываем стоимость таможенной ставки
	chinaCustomsСostPrice, exchangeRateYuanToRub, exchangeRateEuroToRub, err := calculateChinaCustomsСostPrice(typeAuto, priceYuan, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости таможенной ставки китайского авто: %v\n", err)
		return nil, err
	}
	result["chinaCustomsСostPrice"] = RoundUpTo10000(chinaCustomsСostPrice)

	result["exchangeRateYuanToRub"] = math.Round(exchangeRateYuanToRub*100) / 100
	result["exchangeRateEuroToRub"] = math.Round(exchangeRateEuroToRub*100) / 100

	// Рассчитываем стоимость утильсбора
	chinaRecyclingCollection, err := calculateChinaRecyclingCollection(typeAuto, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости утильсбора китайского авто: %v\n", err)
		return nil, err
	}
	result["chinaRecyclingCollection"] = RoundUpTo10000(chinaRecyclingCollection)

	// Получаем стоимость логистики из Китая
	logisticChinaPrice, err := internal.GetValueFromSheet(srv, "Китай", "B2")
	if err != nil {
		return nil, err
	}

	chinaPriceAutoRub := priceYuan * exchangeRateYuanToRub
	result["chinaPriceAutoRub"] = RoundUpTo10000(chinaPriceAutoRub)

	// Рассчитываем стоимость логистики из Китая в рублях
	chinaLogisticPriceRub := logisticChinaPrice * exchangeRateYuanToRub
	result["chinaLogisticPriceRub"] = RoundUpTo10000(chinaLogisticPriceRub)

	result["chinaLogisticsDestination"] = math.Ceil(logisticsDestination)

	// Получаем комиссию
	chinaCommission, err := internal.GetValueFromSheet(srv, "Китай", "C3")
	if err != nil {
		return nil, err
	}
	result["chinaCommission"] = chinaCommission

	// Получаем стоимость оформления документов
	chinaDocumentsPrice, err := internal.GetValueFromSheet(srv, "Китай", "E3")
	if err != nil {
		return nil, err
	}
	result["chinaDocumentsPrice"] = chinaDocumentsPrice

	// Рассчитываем итоговую стоимость
	chinaResultPrice := RoundUpTo10000(chinaCustomsСostPrice) + RoundUpTo10000(chinaRecyclingCollection) + RoundUpTo10000(chinaPriceAutoRub) + RoundUpTo10000(chinaLogisticPriceRub) + chinaCommission + math.Ceil(logisticsDestination) + chinaDocumentsPrice
	result["chinaResultPrice"] = chinaResultPrice

	// Возвращаем map с результатами
	return result, nil
}
