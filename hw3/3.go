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

	"./domain"
	"./generator" //nolint:typecheck
)

//func write(candles []domain.Candle, period domain.CandlePeriod) {
//	var file *os.File
//	if period == domain.CandlePeriod1m {
//		file, _ = os.OpenFile("candles_1m.csv",
//			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	} else if period == domain.CandlePeriod2m {
//		file, _ = os.OpenFile("candles_2m.csv",
//			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	} else if period == domain.CandlePeriod10m {
//		file, _ = os.OpenFile("candles_10m.csv",
//			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	} else {
//		log.Fatalf("Inccorect period on write. exit.")
//	}
//	defer file.Close()
//	for _, candle := range candles {
//		str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
//			candle.Ticker, candle.Period, candle.Open,
//			candle.Close, candle.Low, candle.High)
//		file.WriteString(str)
//	}
//	file.WriteString("---------------\n")
//}

func write(candlesChan <-chan []domain.Candle, period domain.CandlePeriod, wg *sync.WaitGroup, ctx context.Context) {
	fmt.Printf("in print")
	defer wg.Done()
	var file *os.File
	for {
		select {
		case candles := <-candlesChan:
			fmt.Printf("Get fo print: %+v", candles)
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
			for _, candle := range candles {
				str := fmt.Sprintf("%s,%s,%f,%f,%f,%f\n",
					candle.Ticker, candle.Period, candle.Open,
					candle.Close, candle.Low, candle.High)
				file.WriteString(str)
			}
			file.Close()
		case <-ctx.Done():
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
}

func FullUpdateCandle(candles []domain.Candle, newPrice domain.Price, indexCandle int) {
	candles[indexCandle].Open = newPrice.Value
	candles[indexCandle].Close = newPrice.Value
	candles[indexCandle].High = newPrice.Value
	candles[indexCandle].Low = newPrice.Value
	candles[indexCandle].TS = newPrice.TS
}

func CandleWorker(in <-chan domain.Price, ctx context.Context, period domain.CandlePeriod, wg *sync.WaitGroup) (<-chan domain.Price, <-chan []domain.Candle) {
	out := make(chan domain.Price)
	outPrint := make(chan []domain.Candle)

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
					//write(candles, period)
					outPrint <- candles
					candles = nil
					candles = make([]domain.Candle, 0)
					candles = append(candles, domain.NewCandle(newPrice.Ticker, period, newPrice.TS, newPrice.Value))
					indexCandle = len(candles) - 1
					//FullUpdateCandle(candles, newPrice, indexCandle)
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
	oneMinPrice, oneMinCandle := CandleWorker(prices, ctx, domain.CandlePeriod1m, &wg)
	go write(oneMinCandle, domain.CandlePeriod1m, &wg, ctx)
	twoMinPrice, twoMinCandle := CandleWorker(oneMinPrice, ctx, domain.CandlePeriod2m, &wg)
	go write(twoMinCandle, domain.CandlePeriod2m, &wg, ctx)
	tenMinPrice, tenMinCandle := CandleWorker(twoMinPrice, ctx, domain.CandlePeriod10m, &wg)
	go write(tenMinCandle, domain.CandlePeriod10m, &wg, ctx)

	//<-tenMinPrice
	for price := range tenMinPrice {
		//logger.Infof("prices %d: %+v", i, <-prices)
		fmt.Printf("prices: %+v\n", price)
	}
	<-sign
	fmt.Printf("Out\n")
	wg.Wait()
	cancel()
}
