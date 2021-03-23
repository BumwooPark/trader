//go:generate mockgen -source=./store.go -destination=./mock/mock_store.go
package domain

import (
	"context"
	"time"
)

type Store interface {
	Order(code string, ot OrderType, size int64, price float64) error
	Cancel(id string) error
	LoadHistory(ctx context.Context, d time.Duration) ([]Candle, error)
	LoadTick(ctx context.Context) (<-chan Tick, error)
	Uid() string
	Code() string
}
