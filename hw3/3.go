package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"hw3/domain"
	"hw3/generator"
	// log "hw3/github.com/sirupsen/logrus"
)

func write(ctx context.Context, candlesChan <-chan []domain.Candle, period domain.CandlePeriod, wg *sync.WaitGroup) {
	defer wg.Done()
	var file *os.File
	for {
		select {
		case candles := <-candlesChan:
			// fmt.Printf("Get fo print: %+v", candles)
			switch period {
			case domain.CandlePeriod1m:
				fmt.Println("Print to candles_1m.csv")
				file, _ = os.OpenFile("candles_1m.csv",
					os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			case domain.CandlePeriod2m:
				fmt.Println("Print to candles_2m.csv")
				file, _ = os.OpenFile("candles_2m.csv",
					os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			case domain.CandlePeriod10m:
				fmt.Println("Print to candles_10m.csv")
				file, _ = os.OpenFile("candles_10m.csv",
					os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			default:
				fmt.Println("error period write")
			}
			for _, candle := range candles {
				str := fmt.Sprintf("%s,%s,%f,%f,%f,%f,%s\n",
					candle.Ticker, candle.Period, candle.Open,
					candle.Close, candle.Low, candle.High, candle.TS)
				_, err := file.WriteString(str)
				if err != nil {
					fmt.Printf("Error write to file")
				}
			}
			file.Close()
		case <-ctx.Done():
			fmt.Println("Print out")
			return
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
	candles[indexCandle].TS = newPrice.TS
}

func FullUpdateCandle(candles []domain.Candle, newPrice domain.Price, indexCandle int) {
	candles[indexCandle].Open = newPrice.Value
	candles[indexCandle].Close = newPrice.Value
	candles[indexCandle].High = newPrice.Value
	candles[indexCandle].Low = newPrice.Value
	candles[indexCandle].TS = newPrice.TS
}

func CandleWorker(ctx context.Context, in <-chan domain.Price, period domain.CandlePeriod, wg *sync.WaitGroup) (<-chan domain.Price, <-chan []domain.Candle) {
	out := make(chan domain.Price)
	outPrint := make(chan []domain.Candle)

	wg.Add(1)
	go func() {
		defer wg.Done()
		candles := make([]domain.Candle, 0)
		for {
			select {
			case <-ctx.Done():
				outPrint <- candles
				close(outPrint)
				close(out)
				fmt.Println("CandleWorker end")
				return
			case newPrice := <-in:
				// fmt.Printf("Get from chan: %+v\n", newPrice)
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
					// write(candles, period)
					outPrint <- candles
					candles = nil
					candles = append(candles, domain.NewCandle(newPrice.Ticker, period, newPrice.TS, newPrice.Value))
					indexCandle = len(candles) - 1
					// FullUpdateCandle(candles, newPrice, indexCandle)
				} else {
					UpdateCandle(candles, newPrice, indexCandle)
				}
				out <- newPrice
			}
		}
	}()
	return out, outPrint
}

var tickers = []string{"AAPL", "SBER", "NVDA", "TSLA"}

func startGoroutines(ctx context.Context, price <-chan domain.Price, wg *sync.WaitGroup) {
	wg.Add(3)
	defer wg.Done()
	oneMinPrice, oneMinCandle := CandleWorker(ctx, price, domain.CandlePeriod1m, wg)
	go write(ctx, oneMinCandle, domain.CandlePeriod1m, wg)
	twoMinPrice, twoMinCandle := CandleWorker(ctx, oneMinPrice, domain.CandlePeriod2m, wg)
	go write(ctx, twoMinCandle, domain.CandlePeriod2m, wg)
	tenMinPrice, tenMinCandle := CandleWorker(ctx, twoMinPrice, domain.CandlePeriod10m, wg)
	go write(ctx, tenMinCandle, domain.CandlePeriod10m, wg)
	for i := 0; i <= 100; i++ {
		// logger.Infof("prices %+v", price)
		fmt.Printf("prices %d, %+v\n", i, <-tenMinPrice)
	}
	// <-tenMinPrice
	<-ctx.Done()
	fmt.Println("end starter")
}

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
	prices := pg.Prices(ctx)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go startGoroutines(ctx, prices, &wg)

	<-sign
	close(sign)
	cancel()
	fmt.Printf("Out\n")
	wg.Wait()
}
