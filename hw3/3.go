package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"tinnkoff_hw/hw3/domain"
	"tinnkoff_hw/hw3/generator"
)

func write(candles []domain.Candle, file *os.File) {
	fmt.Printf("in print")
	fmt.Printf("Get fo print: %+v\n", candles)

	for _, candle := range candles {
		str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
			candle.Ticker, candle.Period, candle.Open,
			candle.Close, candle.Low, candle.High)
		_, err := file.WriteString(str)
		if err != nil {
			log.Fatalf("Error writting to file.\n")
		}
	}
}

func UpdateCandle(candles []domain.Candle, newPrice domain.Price, indexCandle int) {
	if candles[indexCandle].Low > newPrice.Value {
		candles[indexCandle].Low = newPrice.Value
	} else if candles[indexCandle].High < newPrice.Value {
		candles[indexCandle].High = newPrice.Value
	}
	candles[indexCandle].Close = newPrice.Value
}

func emptyFunc(ctx context.Context, in <-chan domain.Price) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-in:
			continue
		}
	}
}

func CandleWorker(ctx context.Context, in <-chan domain.Price, period domain.CandlePeriod, wg *sync.WaitGroup) <-chan domain.Price {
	out := make(chan domain.Price)
	var file *os.File

	switch period {
	case domain.CandlePeriod1m:
		file, _ = os.Create("candles_1m.csv")
	case domain.CandlePeriod2m:
		file, _ = os.Create("candles_2m.csv")
	case domain.CandlePeriod10m:
		file, _ = os.Create("candles_10m.csv")
	default:
		log.Fatalf("Inccorect period on write. exit.")
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		candles := make([]domain.Candle, 0)
		for {
			select {
			case <-ctx.Done():
				write(candles, file)
				return
			case newPrice := <-in:
				indexCandle := -1
				for i, candle := range candles {
					if candle.Ticker == newPrice.Ticker {
						indexCandle = i
						break
					}
				}
				if indexCandle == -1 {
					candles = append(candles, domain.NewCandle(newPrice.Ticker, period, newPrice.TS, newPrice.Value))
					indexCandle = len(candles) - 1
					out <- newPrice
					break
				}
				prevPeriod, _ := domain.PeriodTS(period, candles[indexCandle].TS)
				currPeriod, _ := domain.PeriodTS(period, newPrice.TS)
				if prevPeriod != currPeriod {
					write(candles, file)
					candles = make([]domain.Candle, 0)
					candles = append(candles, domain.NewCandle(newPrice.Ticker, period, newPrice.TS, newPrice.Value))
					indexCandle = len(candles) - 1
				} else {
					UpdateCandle(candles, newPrice, indexCandle)
				}
				out <- newPrice
			}
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

	// logger.Info("start prices generator...")
	fmt.Println("start prices generator...")
	prices := pg.Prices(ctx)
	wg := sync.WaitGroup{}

	oneMinPrice := CandleWorker(ctx, prices, domain.CandlePeriod1m, &wg)
	twoMinPrice := CandleWorker(ctx, oneMinPrice, domain.CandlePeriod2m, &wg)
	tenMinPrice := CandleWorker(ctx, twoMinPrice, domain.CandlePeriod10m, &wg)

	go emptyFunc(ctx, tenMinPrice)
	<-sign
	fmt.Printf("Out\n")
	cancel()
	wg.Wait()
}
