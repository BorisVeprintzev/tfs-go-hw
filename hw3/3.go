package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"tinnkoff_hw/hw3/domain"
	"tinnkoff_hw/hw3/generator"

	log "github.com/sirupsen/logrus"
)

func UpdateCandle(candles []domain.Candle, newPrice domain.Price, indexCandle int) {
	if candles[indexCandle].Low > newPrice.Value {
		candles[indexCandle].Low = newPrice.Value
	} else if candles[indexCandle].High < newPrice.Value {
		candles[indexCandle].High = newPrice.Value
	}
	candles[indexCandle].Close = newPrice.Value
}

func OneMinuteCandle(price <-chan domain.Price, wg *sync.WaitGroup) <-chan domain.Candle {
	out := make(chan domain.Candle)

	wg.Add(1)
	go func() {
		defer close(out)
		defer wg.Done()
		candles := make([]domain.Candle, 0)
		for newPrice := range price {
			log.Info(fmt.Sprintf("Get new Price %+v", newPrice))
			indexCandle := -1
			for i, candle := range candles {
				if candle.Ticker == newPrice.Ticker {
					indexCandle = i
					break
				}
			}
			if indexCandle == -1 {
				candles = append(candles, domain.NewCandle(newPrice.Ticker, domain.CandlePeriod1m, newPrice.TS, newPrice.Value))
				indexCandle = len(candles) - 1
			}
			prevPeriod, _ := domain.PeriodTS(domain.CandlePeriod1m, candles[indexCandle].TS)
			currPeriod, _ := domain.PeriodTS(domain.CandlePeriod1m, newPrice.TS)
			if prevPeriod != currPeriod {
				for _, candle := range candles {
					out <- candle
				}
				candles = make([]domain.Candle, 0)
				candles = append(candles, domain.NewCandle(newPrice.Ticker, domain.CandlePeriod1m, newPrice.TS, newPrice.Value))
			} else {
				UpdateCandle(candles, newPrice, indexCandle)
			}
		}
		for _, candle := range candles {
			out <- candle
		}
	}()
	return out
}

func WriteMinute(in <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	log.Info("Write 1 minute")
	file, err := os.Create("candles_1m.csv")
	if err != nil {
		log.Error("Error create fail")
	}

	out := make(chan domain.Candle)

	wg.Add(1)
	go func() {
		defer close(out)
		defer wg.Done()
		for candle := range in {
			str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
				candle.Ticker, candle.Period, candle.Open,
				candle.Close, candle.Low, candle.High)
			_, err := file.WriteString(str)
			if err != nil {
				log.Fatalf("Error writting to file.\n")
			}
			out <- candle
		}
	}()
	return out
}

func WriteTwoMinute(in <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	log.Info("Write 2 minute")
	file, err := os.Create("candles_2m.csv")
	if err != nil {
		log.Error("Error create file")
	}

	out := make(chan domain.Candle)

	wg.Add(1)
	go func() {
		defer close(out)
		defer wg.Done()
		for candle := range in {
			str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
				candle.Ticker, candle.Period, candle.Open,
				candle.Close, candle.Low, candle.High)
			_, err := file.WriteString(str)
			if err != nil {
				log.Fatalf("Error writting to file.\n")
			}
			out <- candle
		}
	}()
	return out
}

func WriteTenMinute(in <-chan domain.Candle, wg *sync.WaitGroup) {
	log.Info("Write 10 minute")
	file, err := os.Create("candles_10m.csv")
	if err != nil {
		log.Error("Error create file")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for candle := range in {
			str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
				candle.Ticker, candle.Period, candle.Open,
				candle.Close, candle.Low, candle.High)
			_, err := file.WriteString(str)
			if err != nil {
				log.Fatalf("Error writting to file.\n")
			}
		}
	}()
}

func FullUpdateCandle(candles []domain.Candle, newCandle domain.Candle, idxCandles int) {
	candles[idxCandles].TS = newCandle.TS
	if newCandle.High > candles[idxCandles].High {
		candles[idxCandles].High = newCandle.High
	}
	if newCandle.Low < candles[idxCandles].Low {
		candles[idxCandles].Low = newCandle.Low
	}
	candles[idxCandles].Close = newCandle.Close
}

func TwoMinuteCandle(in <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	out := make(chan domain.Candle)

	wg.Add(1)
	go func() {
		defer close(out)
		defer wg.Done()
		twoMinuteCandles := make([]domain.Candle, 0)
		for oneMinuteCandle := range in {
			log.Info("Get new candle. 2 minute")
			indexCandle := -1
			for idx, candle := range twoMinuteCandles {
				if candle.Ticker == oneMinuteCandle.Ticker {
					indexCandle = idx
				}
			}
			if indexCandle == -1 {
				twoMinuteCandles = append(twoMinuteCandles, domain.FullNewCandle(oneMinuteCandle.Ticker, domain.CandlePeriod2m,
					oneMinuteCandle.TS, oneMinuteCandle.Open, oneMinuteCandle.Close, oneMinuteCandle.High, oneMinuteCandle.Low))
				indexCandle = len(twoMinuteCandles) - 1
			}
			prevPeriod, _ := domain.PeriodTS(domain.CandlePeriod2m, twoMinuteCandles[indexCandle].TS)
			currPeriod, _ := domain.PeriodTS(domain.CandlePeriod2m, oneMinuteCandle.TS)
			if prevPeriod != currPeriod {
				for _, candle := range twoMinuteCandles {
					out <- candle
				}
				twoMinuteCandles = make([]domain.Candle, 0)
				twoMinuteCandles = append(twoMinuteCandles, domain.FullNewCandle(oneMinuteCandle.Ticker, domain.CandlePeriod2m,
					oneMinuteCandle.TS, oneMinuteCandle.Open, oneMinuteCandle.Close, oneMinuteCandle.High, oneMinuteCandle.Low))
				indexCandle = len(twoMinuteCandles) - 1
			} else {
				FullUpdateCandle(twoMinuteCandles, oneMinuteCandle, indexCandle)
			}
		}
		for _, candle := range twoMinuteCandles {
			out <- candle
		}
	}()
	return out
}

func TenMinuteCandle(in <-chan domain.Candle, wg *sync.WaitGroup) <-chan domain.Candle {
	out := make(chan domain.Candle)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		tenMinuteCandles := make([]domain.Candle, 0)
		for twoMinuteCandle := range in {
			log.Info("Get new candle. 10 min")
			indexCandle := -1
			for idx, candle := range tenMinuteCandles {
				if candle.Ticker == twoMinuteCandle.Ticker {
					indexCandle = idx
				}
			}
			if indexCandle == -1 {
				tenMinuteCandles = append(tenMinuteCandles, domain.FullNewCandle(twoMinuteCandle.Ticker, domain.CandlePeriod10m,
					twoMinuteCandle.TS, twoMinuteCandle.Open, twoMinuteCandle.Close, twoMinuteCandle.High, twoMinuteCandle.Low))
				indexCandle = len(tenMinuteCandles) - 1
			}
			prevPeriod, _ := domain.PeriodTS(domain.CandlePeriod10m, tenMinuteCandles[indexCandle].TS)
			currPeriod, _ := domain.PeriodTS(domain.CandlePeriod10m, twoMinuteCandle.TS)
			if prevPeriod != currPeriod {
				for _, candle := range tenMinuteCandles {
					out <- candle
				}
				tenMinuteCandles = make([]domain.Candle, 0)
				tenMinuteCandles = append(tenMinuteCandles, domain.FullNewCandle(twoMinuteCandle.Ticker, domain.CandlePeriod10m,
					twoMinuteCandle.TS, twoMinuteCandle.Open, twoMinuteCandle.Close, twoMinuteCandle.High, twoMinuteCandle.Low))
				indexCandle = len(tenMinuteCandles) - 1
			} else {
				FullUpdateCandle(tenMinuteCandles, twoMinuteCandle, indexCandle)
			}
		}
		for _, candle := range tenMinuteCandles {
			out <- candle
		}
	}()
	return out
}

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

func main() {
	// logger := log.New()
	fmt.Println("Begin work")
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(context.Background())

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	log.Info("start prices generator...")
	prices := pg.Prices(ctx)
	wg := sync.WaitGroup{}

	minuteCandleWrite := OneMinuteCandle(prices, &wg)
	minuteCandle := WriteMinute(minuteCandleWrite, &wg)
	twoMinuteCandleWrite := TwoMinuteCandle(minuteCandle, &wg)
	twoMinuteCandle := WriteTwoMinute(twoMinuteCandleWrite, &wg)
	tenMinuteWrite := TenMinuteCandle(twoMinuteCandle, &wg)
	WriteTenMinute(tenMinuteWrite, &wg)

	log.Info("I'm here")
	log.Info(fmt.Sprintf("Get signal %s", <-sign))
	cancel()
	wg.Wait()
}
