package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gobenpark/proto/stock"
	"github.com/gobenpark/trader/container"
	"github.com/gobenpark/trader/event"
	"github.com/gobenpark/trader/order"
	"github.com/gobenpark/trader/position"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type store struct {
	name      string
	uid       string
	cli       stock.StockClient
	orderList map[string]bool
	isTest    bool
}

func NewStore(name string, isTest bool) *store {
	cli, err := stock.NewSocketClient(context.Background(), "localhost:50051")
	if err != nil {
		panic(err)
	}

	return &store{name, uuid.NewV4().String(), cli, map[string]bool{}, isTest}
}

func (s *store) LoadHistory(ctx context.Context, du time.Duration) ([]container.Candle, error) {
	code := "KRW-BTC"
	r, err := s.cli.Chart(context.Background(), &stock.ChartRequest{
		Code: code,
		To:   nil,
	})
	if err != nil {
		return nil, err
	}
	var d []container.Candle
	for _, i := range r.GetData() {
		d = append(d, container.Candle{
			Code:   code,
			Low:    i.GetLow(),
			High:   i.GetHigh(),
			Open:   i.GetOpen(),
			Close:  i.GetClose(),
			Volume: i.GetVolume(),
			Date:   i.GetDate().AsTime(),
		})
	}

	sort.Slice(d, func(i, j int) bool {
		return d[i].Date.Before(d[j].Date)
	})

	return d, nil
}

func (s *store) LoadTick(ctx context.Context) (<-chan container.Tick, error) {
	ch := make(chan container.Tick, 1)

	r, err := s.cli.TickStream(ctx, &stock.TickRequest{Codes: ""})
	if err != nil {
		return nil, err
	}
	go func(tch chan container.Tick) {
	Done:
		for {
			select {
			case <-ctx.Done():
				break Done
			default:
				msg, err := r.Recv()
				if err != nil {
					st, _ := status.FromError(err)
					if st.Code() == codes.Unimplemented || st.Code() == codes.Unavailable {
						break
					}
					log.Err(err).Send()
				}
				tch <- container.Tick{
					Code:   "",
					Date:   msg.GetDate().AsTime(),
					Price:  msg.GetPrice(),
					Volume: msg.GetVolume(),
				}
			}
		}
	}(ch)

	return ch, nil
}

func (s *store) Order(o *order.Order) error {
	switch o.OType {
	case order.Buy:
		if !s.isTest {
			_, err := s.cli.Buy(context.Background(), &stock.BuyRequest{
				Code:       o.Code,
				Otype:      stock.OrderType_LimitOrder,
				Volume:     float64(o.Size),
				Price:      o.Price,
				Identifier: o.UUID,
			})
			if err != nil {
				return err
			}
		}
	case order.Sell:
		if !s.isTest {
			_, err := s.cli.Sell(context.Background(), &stock.SellRequest{
				Code:       o.Code,
				Otype:      stock.OrderType_LimitOrder,
				Volume:     float64(o.Size),
				Price:      o.Price,
				Identifier: o.UUID,
			})
			if err != nil {
				return err
			}
		}
	}
	s.orderList[o.UUID] = true
	return nil
}

func (s *store) Cancel(id string) error {
	if !s.isTest {
		if _, err := s.cli.CancelOrder(context.Background(), &stock.CancelOrderRequest{Id: id}); err != nil {
			return err
		}
	}
	s.orderList[id] = false

	return nil
}

func (s *store) Uid() string {
	return s.uid
}

func (s *store) Cash() int64 {
	res, err := s.cli.Position(context.Background(), &emptypb.Empty{})
	if err != nil {
		return 0
	}
	for _, i := range res.GetPositions() {
		if i.GetCode() == "KRW" {
			return int64(i.Amount)
		}
	}
	return 0
}

func (s *store) Commission() float64 {
	return 0.0005
}

func (s *store) Positions() []position.Position {
	res, err := s.cli.Position(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil
	}

	var p []position.Position
	for _, i := range res.GetPositions() {
		p = append(p, position.Position{
			Code:      i.GetCode(),
			Size:      int64(i.GetAmount()),
			Price:     i.GetPrice(),
			CreatedAt: i.GetDate().AsTime(),
		})
	}
	return p
}

func (s *store) AllCodes() (map[string]string, error) {
	codes := map[string]string{}
	res, err := s.cli.AllMarkets(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	for _, i := range res.GetMarkets() {
		codes[i.GetCode()] = i.GetName()
	}
	return codes, nil
}

func (s *store) OrderState(ctx context.Context) (<-chan event.OrderEvent, error) {
	ch := make(chan event.OrderEvent)
	ticker := time.NewTicker(time.Second * 3)

	go func() {
		defer close(ch)
	Done:
		for {
			select {
			case <-ticker.C:
				for i := range s.orderList {
					resp, err := s.cli.OrderInfo(ctx, &stock.OrderInfoRequest{Id: i})
					if err != nil {
						fmt.Println(err)
						continue
					}
					ch <- event.OrderEvent{
						Event: event.Event{
							EventType: "order",
							Message:   "order event rise",
						},
						State: resp.GetOrder().GetState(),
						Oid:   i,
					}
				}
			case <-ctx.Done():
				break Done

			}
		}
	}()

	return ch, nil
}

func (s *store) OrderInfo(id string) (*order.Order, error) {
	re, err := s.cli.OrderInfo(context.Background(), &stock.OrderInfoRequest{Id: id})
	if err != nil {
		return nil, err
	}

	o := &order.Order{
		Code:       re.GetOrder().GetCode(),
		UUID:       id,
		Size:       int64(re.GetOrder().GetVolume()),
		Price:      re.GetOrder().GetPrice(),
		CreatedAt:  re.GetOrder().GetCreatedAt().AsTime(),
		ExecutedAt: time.Time{},
	}

	switch re.GetOrder().GetState() {
	case "wait":
		o.Submit()
	case "cancel":
		o.Cancel()
	case "done":
		o.Complete()
	}

	return o, nil
}
