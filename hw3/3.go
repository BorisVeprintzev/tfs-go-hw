package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"./domain"
	"./generator"
)

type ProtectedSliceCandles struct {
	mx      sync.Mutex
	Candles []domain.Candle
}

func (s *ProtectedSliceCandles) Lock() {
	s.mx.Lock()
}

func (s *ProtectedSliceCandles) Unlock() {
	s.mx.Unlock()
}

func NewProtectedSliceCandles() ProtectedSliceCandles {
	new := ProtectedSliceCandles{
		Candles: make([]domain.Candle, 4),
	}
	new.Unlock()
	return new
}

func write(candles []domain.Candle, period domain.CandlePeriod) {
	var file *os.File
	if period == domain.CandlePeriod1m {
		file, _ = os.OpenFile("candles_1m.csv",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else if period == domain.CandlePeriod2m {
		file, _ = os.OpenFile("candles_2m.csv",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else if period == domain.CandlePeriod10m {
		file, _ = os.OpenFile("candles_10m.csv",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		log.Fatalf("Inccorect period on write. exit.")
	}
	defer file.Close()
	for _, candle := range candles {
		str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
			candle.Ticker, candle.TS.String(), candle.Open,
			candle.Close, candle.Low, candle.High)
		file.WriteString(str)
	}
	file.WriteString("---------------\n")
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

func CandleWorker(in <-chan domain.Price, ctx context.Context, period domain.CandlePeriod, wg *sync.WaitGroup) <-chan domain.Price {
	out := make(chan domain.Price)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		fmt.Println("ya tyt")
		candles := make([]domain.Candle, 0)
		for {
			select {
			case <-ctx.Done():
				write(candles, period)
				return
			case <-in:
				newPrice := <-in
				// fmt.Printf("%+v\n, %d\n", newPrice, runtime.NumGoroutine())
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
					write(candles, period) // need print only one candle
					FullUpdateCandle(candles, newPrice, indexCandle)
					out <- newPrice
					break
				}
				UpdateCandle(candles, newPrice, indexCandle)
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

	oneMinPrice := CandleWorker(prices, ctx, domain.CandlePeriod1m, &wg)
	twoMinPrice := CandleWorker(oneMinPrice, ctx, domain.CandlePeriod2m, &wg)
	final := CandleWorker(twoMinPrice, ctx, domain.CandlePeriod10m, &wg)

	wg.Wait()
	for price := range final {
		//logger.Infof("prices %d: %+v", i, <-prices)
		fmt.Printf("prices: %+v\n", price)
	}
	time.Sleep(2)
	cancel()
}
