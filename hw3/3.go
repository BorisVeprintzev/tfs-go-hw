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

func write(candlesChan <-chan []domain.Candle, period domain.CandlePeriod, wg *sync.WaitGroup) {
	fmt.Printf("in print")
	defer wg.Done()
	var file *os.File
	for candles := range candlesChan {
		fmt.Printf("Get fo print: %+v", candles)
		if period == domain.CandlePeriod1m {
			file, _ = os.Create("candles_1m.csv")
		} else if period == domain.CandlePeriod2m {
			file, _ = os.Create("candles_2m.csv")
		} else if period == domain.CandlePeriod10m {
			file, _ = os.Create("candles_10m.csv")
		} else {
			log.Fatalf("Inccorect period on write. exit.")
		}
		for _, candle := range candles {
			str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
				candle.Ticker, candle.Period, candle.Open,
				candle.Close, candle.Low, candle.High)
			file.WriteString(str)
		}
		file.Close()
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

func FullUpdateCandle(candles []domain.Candle, newPrice domain.Price, indexCandle int) {
	candles[indexCandle].Open = newPrice.Value
	candles[indexCandle].Close = newPrice.Value
	candles[indexCandle].High = newPrice.Value
	candles[indexCandle].Low = newPrice.Value
	candles[indexCandle].TS = newPrice.TS
}

func CandleWorker(ctx context.Context, in <-chan domain.Price, period domain.CandlePeriod, wg *sync.WaitGroup) <-chan domain.Price {
	out := make(chan domain.Price)
	outPrint := make(chan []domain.Candle)

	go write(outPrint, period, wg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		defer close(outPrint)
		candles := make([]domain.Candle, 0)
		for {
			select {
			case <-ctx.Done():
				outPrint <- candles
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
					outPrint <- candles
					candles = nil
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
	//logger := log.New()
	fmt.Println("Begin work")
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(context.Background())

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	//logger.Info("start prices generator...")
	fmt.Println("start prices generator...")
	prices := pg.Prices(ctx)
	wg := sync.WaitGroup{}

	wg.Add(3)
	oneMinPrice := CandleWorker(ctx, prices, domain.CandlePeriod1m, &wg)
	twoMinPrice := CandleWorker(ctx, oneMinPrice, domain.CandlePeriod2m, &wg)
	tenMinPrice := CandleWorker(ctx, twoMinPrice, domain.CandlePeriod10m, &wg)

	//<-tenMinPrice
	for price := range tenMinPrice {
		//logger.Infof("prices %d: %+v", i, <-prices)
		fmt.Printf("prices: %+v\n", price)
	}
	<-sign
	fmt.Printf("Out\n")
	cancel()
	wg.Wait()
}
