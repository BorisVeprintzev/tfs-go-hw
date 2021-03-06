package generator

import (
	"context"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	"tinnkoff_hw/hw3/domain"
)

type Config struct {
	Factor  float64
	Delay   time.Duration
	Tickers []string
}

func NewPricesGenerator(cfg Config) *PricesGenerator {
	return &PricesGenerator{
		factor:  cfg.Factor,
		delay:   cfg.Delay,
		tickers: cfg.Tickers,
	}
}

type PricesGenerator struct {
	factor  float64
	delay   time.Duration
	tickers []string
}

func (g *PricesGenerator) Prices(ctx context.Context) <-chan domain.Price {
	prices := make(chan domain.Price)

	startTime := time.Now()
	go func() {
		defer close(prices)

		ticker := time.NewTicker(g.delay)
		for {
			select {
			case <-ctx.Done():
				log.Info("Prices generate done")
				return
			case <-ticker.C:
				currentTime := time.Now()
				delta := float64(currentTime.Unix() - startTime.Unix())
				ts := time.Unix(int64(float64(currentTime.Unix())+delta*g.factor), 0)

				for idx, ticker := range g.tickers {
					min := float64((idx + 1) * 100)
					max := min + 20
					tmp := domain.Price{
						Ticker: ticker,
						Value:  min + rand.Float64()*(max-min),
						TS:     ts,
					}
					// fmt.Printf("Send to chan: %+v\n", tmp)
					prices <- tmp
				}
			}
		}
	}()

	return prices
}
