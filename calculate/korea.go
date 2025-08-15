package calculate

import (
	"fmt"
	"log"
	"math"
	"time"

	"go-wannacar-calculator/internal"

	"google.golang.org/api/sheets/v4"
)

func RoundUpTo10000(value float64) float64 {
	return math.Ceil(value/100) * 100
}

// Функция для расчета утильсбора
func calculateKoreaRecyclingCollection(typeAuto string, carAge int, engineVolume float64, srv *sheets.Service) (float64, error) {
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
func calculateKoreaCustomsСostPrice(typeAuto string, priceWon float64, carAge int, engineVolume float64, srv *sheets.Service) (float64, float64, float64, error) {
	// Получаем значение из ячейки "Курс!B4" (курс вон к рублю)
	exchangeRateWonToRub, err := internal.GetValueFromSheet(srv, "Курс", "B4")
	if err != nil {
		return 0, 0, 0, err
	}
	log.Printf("₩: %f ₽\n", exchangeRateWonToRub)

	// Получаем значение из ячейки "Курс!B3" (курс евро к рублю)
	exchangeRateEuroToRub, err := internal.GetValueFromSheet(srv, "Курс", "B3")
	if err != nil {
		return 0, 0, 0, err
	}
	log.Printf("€: %f ₽\n", exchangeRateEuroToRub)

	log.Printf("Стоимость авто в вонах: %.2f ₩\n", priceWon)
	log.Printf("Возраст авто: %d\n", carAge)
	log.Printf("Объем двигателя в см³: %.2f см³\n", engineVolume)

	// Преобразуем стоимость авто в евро
	priceEuro := (priceWon * exchangeRateWonToRub) / exchangeRateEuroToRub
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

			customsСostPrice = priceWon * exchangeRateWonToRub * (rate / 100)
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

		customsСostPrice = priceWon * exchangeRateWonToRub * (rate / 100)
		log.Printf("Таможенная ставка рассчитанная от стоимости электромобиля: %.2f ₽\n", customsСostPrice)

	}

	return customsСostPrice, exchangeRateWonToRub, exchangeRateEuroToRub, nil
}

func CalculateKoreaResult(typeAuto string, priceWon float64, yearOfManufacture int, engineVolume float64, logisticsDestination float64, srv *sheets.Service) (map[string]float64, error) {
	// Инициализируем map для хранения результатов
	result := make(map[string]float64)

	// Получаем текущий год
	currentYear := time.Now().Year()

	// Рассчитываем возраст авто
	carAge := currentYear - yearOfManufacture

	// Рассчитываем стоимость таможенной ставки
	koreaCustomsСostPrice, exchangeRateWonToRub, exchangeRateEuroToRub, err := calculateKoreaCustomsСostPrice(typeAuto, priceWon, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости таможенной ставки корейского авто: %v\n", err)
		return nil, err
	}
	result["koreaCustomsСostPrice"] = RoundUpTo10000(koreaCustomsСostPrice)

	result["exchangeRateWonToRub"] = math.Round(exchangeRateWonToRub*1000*100) / 100
	result["exchangeRateEuroToRub"] = math.Round(exchangeRateEuroToRub*100) / 100

	// Рассчитываем стоимость утильсбора
	koreaRecyclingCollection, err := calculateKoreaRecyclingCollection(typeAuto, carAge, engineVolume, srv)
	if err != nil {
		log.Printf("Ошибка при расчете стоимости утильсбора корейского авто: %v\n", err)
		return nil, err
	}
	result["koreaRecyclingCollection"] = RoundUpTo10000(koreaRecyclingCollection)

	// Получаем стоимость логистики из Кореи
	logisticKoreaPrice, err := internal.GetValueFromSheet(srv, "Корея", "B2")
	if err != nil {
		return nil, err
	}

	koreaPriceAutoRub := priceWon * exchangeRateWonToRub
	result["koreaPriceAutoRub"] = RoundUpTo10000(koreaPriceAutoRub)

	// Рассчитываем стоимость логистики из Кореи в рублях
	koreaLogisticPriceRub := logisticKoreaPrice * exchangeRateWonToRub
	result["koreaLogisticPriceRub"] = RoundUpTo10000(koreaLogisticPriceRub)

	result["koreaLogisticsDestination"] = math.Ceil(logisticsDestination)

	// Получаем комиссию
	koreaCommission, err := internal.GetValueFromSheet(srv, "Корея", "C3")
	if err != nil {
		return nil, err
	}
	result["koreaCommission"] = koreaCommission

	// Получаем стоимость логистики по Москве
	// koreaLogisticMoscowPrice, err := internal.GetValueFromSheet(srv, "Корея", "D3")
	// if err != nil {
	// 	return nil, err
	// }
	// result["koreaLogisticMoscowPrice"] = koreaLogisticMoscowPrice

	// Получаем стоимость оформления документов
	koreaDocumentsPrice, err := internal.GetValueFromSheet(srv, "Корея", "E3")
	if err != nil {
		return nil, err
	}
	result["koreaDocumentsPrice"] = koreaDocumentsPrice

	// Рассчитываем итоговую стоимость
	koreaResultPrice := RoundUpTo10000(koreaCustomsСostPrice) + RoundUpTo10000(koreaRecyclingCollection) + RoundUpTo10000(koreaPriceAutoRub) + RoundUpTo10000(koreaLogisticPriceRub) + koreaCommission + math.Ceil(logisticsDestination) + koreaDocumentsPrice
	result["koreaResultPrice"] = koreaResultPrice

	// Возвращаем map с результатами
	return result, nil
}
