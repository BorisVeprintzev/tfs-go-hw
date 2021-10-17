package domain //nolint:goimports

import (
	"errors"
	"time"
)

type Price struct {
	Ticker string
	Value  float64
	TS     time.Time
}

var ErrUnknownPeriod = errors.New("unknown period")

type CandlePeriod string

const (
	CandlePeriod1m  CandlePeriod = "1m"
	CandlePeriod2m  CandlePeriod = "2m"
	CandlePeriod10m CandlePeriod = "10m"
)

func PeriodTS(period CandlePeriod, ts time.Time) (time.Time, error) {
	switch period {
	case CandlePeriod1m:
		return ts.Truncate(time.Minute), nil
	case CandlePeriod2m:
		return ts.Truncate(2 * time.Minute), nil
	case CandlePeriod10m:
		return ts.Truncate(10 * time.Minute), nil
	default:
		return time.Time{}, ErrUnknownPeriod
	}
}

type Candle struct {
	Ticker string
	Period CandlePeriod // Интервал
	Open   float64      // Цена открытия
	High   float64      // Максимальная цена
	Low    float64      // Минимальная цена
	Close  float64      // Цена закрытие
	TS     time.Time    // Время начала интервала
}

func NewCandle(ticker string, period CandlePeriod, beginTime time.Time, open float64) Candle {
	return Candle{
		Ticker: ticker,
		Period: period,
		TS:     beginTime,
		Open:   open,
		Close:  open,
		High:   open,
		Low:    open,
	}
}

func FullNewCandle(ticker string, period CandlePeriod, beginTime time.Time, open float64, close float64, high float64, low float64) Candle {
	return Candle{
		Ticker: ticker,
		Period: period,
		TS:     beginTime,
		Open:   open,
		Close:  close,
		High:   high,
		Low:    low,
	}
}
