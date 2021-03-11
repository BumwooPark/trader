package store

//go:generate mockgen -source=./store.go -destination=./mock/mock_store.go

import (
	"context"

	"github.com/gobenpark/proto/stock"
	"github.com/gobenpark/trader/domain"
	"github.com/gogo/protobuf/types"
	"github.com/rs/zerolog/log"
)

type store struct {
	cli stock.StockClient
}

func NewStore() *store {
	cli, err := stock.NewSocketClient(context.Background(), "localhost:50051")
	if err != nil {
		panic(err)
	}

	return &store{cli}
}

func (s *store) LoadHistory(ctx context.Context, code string) ([]domain.Candle, error) {
	r, err := s.cli.Chart(context.Background(), &stock.ChartRequest{
		Code: code,
		To:   nil,
	})
	if err != nil {
		return nil, err
	}
	var d []domain.Candle
	for _, i := range r.GetData() {
		ti, err := types.TimestampFromProto(i.GetDate())
		if err != nil {
			log.Err(err).Send()
		}
		d = append(d, domain.Candle{
			Code:   code,
			Low:    i.GetLow(),
			High:   i.GetHigh(),
			Open:   i.GetOpen(),
			Close:  i.GetClose(),
			Volume: i.GetVolume(),
			Date:   ti,
		})
	}
	return d, nil
}

func (s *store) LoadTick(ctx context.Context, code string) (<-chan domain.Tick, error) {
	ch := make(chan domain.Tick, 1)

	r, err := s.cli.TickStream(ctx, &stock.TickRequest{Codes: code})
	if err != nil {
		return nil, err
	}
	go func(tch chan domain.Tick) {
	Done:
		for {
			select {
			case <-ctx.Done():
				break Done
			default:
				msg, err := r.Recv()
				if err != nil {
					log.Err(err).Send()
				}
				ti, err := types.TimestampFromProto(msg.GetDate())
				if err != nil {
					log.Err(err).Send()
				}
				tch <- domain.Tick{
					Code:   code,
					Date:   ti,
					Price:  msg.GetPrice(),
					Volume: msg.GetVolume(),
				}
			}
		}
	}(ch)

	return ch, nil
}
