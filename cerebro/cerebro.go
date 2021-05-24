/*
 *                     GNU GENERAL PUBLIC LICENSE
 *                        Version 3, 29 June 2007
 *
 *  Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
 *  Everyone is permitted to copy and distribute verbatim copies
 *  of this license document, but changing it is not allowed.
 *
 *                             Preamble
 *
 *   The GNU General Public License is a free, copyleft license for
 * software and other kinds of works.
 */

package cerebro

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gobenpark/trader/broker"
	"github.com/gobenpark/trader/container"
	error2 "github.com/gobenpark/trader/error"
	"github.com/gobenpark/trader/event"
	"github.com/gobenpark/trader/internal/pkg"
	"github.com/gobenpark/trader/market"
	"github.com/gobenpark/trader/store"
	"github.com/gobenpark/trader/strategy"
)

// Cerebro head of trading system
// make all dependency manage
type Cerebro struct {
	//broker Sell/Buy and manage order
	broker *broker.Broker

	//store outter store
	store store.Store

	//Ctx cerebro global context
	Ctx context.Context `json:"ctx"`

	//Cancel cerebro global context cancel
	Cancel context.CancelFunc `json:"cancel"`

	//strategies
	strategies []strategy.Strategy

	//compress compress info map for codes
	compress []CompressInfo

	//TODO: information manager for stock, news, etc.
	information string

	markets map[string]*market.Market

	//strategy.StrategyEngine embedding property for managing user strategy
	strategyEngine *strategy.Engine

	//log in cerebro global logger
	Logger Logger

	// eventEngine engine of management all event
	eventEngine *event.Engine
}

//NewCerebro generate new cerebro with cerebro option
func NewCerebro(opts ...Option) *Cerebro {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Cerebro{
		Ctx:            ctx,
		Cancel:         cancel,
		compress:       []CompressInfo{},
		strategyEngine: &strategy.Engine{},
		markets:        make(map[string]*market.Market),
		eventEngine:    event.NewEventEngine(),
		broker:         broker.NewBroker(),
	}

	for _, opt := range opts {
		opt(c)
	}
	if c.Logger == nil {
		c.Logger = GetLogger()
	}

	return c
}

// orderEventRoutine is stream of order state
// if rise order event then event hub send to subscriber
func (c *Cerebro) orderEventRoutine() {
	ch, err := c.store.OrderState(c.Ctx)
	if err != nil {
		panic(err)
	}

	go func() {
		for i := range ch {
			c.eventEngine.BroadCast(i)
		}
	}()
}

func (c *Cerebro) loadHistoryData() {
	c.store.LoadHistory(c.Ctx, time.Second)
}

//loadRealtimeData
func (c *Cerebro) loadRealtimeData() (<-chan container.Tick, error) {
	c.Logger.Info("start load live data")

	var tick <-chan container.Tick
	if err := pkg.Retry(10, func() error {
		var err error
		tick, err = c.store.LoadTick(c.Ctx)
		if err != nil {
			c.Logger.Warning("try restart store loadTick...")
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return tick, nil
}

// registerEvent is resiter event listener
func (c *Cerebro) registerEvent() {
	c.eventEngine.Register <- c.strategyEngine
	c.eventEngine.Register <- c.broker
}

//Start run cerebro
// first check cerebro validation
// second load from store data
// third other engine setup
func (c *Cerebro) Start() error {
	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGTERM)

	if c.store == nil {
		return error2.ErrStoreNotExists
	}

	c.eventEngine.Start(c.Ctx)
	c.registerEvent()
	c.broker.Store = c.store
	c.strategyEngine.Broker = c.broker

	c.orderEventRoutine()

	select {
	case <-c.Ctx.Done():
		break
	case <-done:
		break
	}
	return nil
}

//Stop all cerebro goroutine and finish
func (c *Cerebro) Stop() error {
	c.Cancel()
	return nil
}
