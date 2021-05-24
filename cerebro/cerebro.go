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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/gobenpark/trader/broker"
	"github.com/gobenpark/trader/container"
	error2 "github.com/gobenpark/trader/error"
	"github.com/gobenpark/trader/event"
	"github.com/gobenpark/trader/internal/pkg"
	"github.com/gobenpark/trader/market"
	"github.com/gobenpark/trader/order"
	"github.com/gobenpark/trader/store"
	"github.com/gobenpark/trader/strategy"
)

// Cerebro head of trading system
// make all dependency manage
type Cerebro struct {
	//Broker buy, sell and manage order
	broker *broker.Broker `validate:"required"`

	store store.Store

	//Ctx cerebro global context
	Ctx context.Context `json:"ctx" validate:"required"`

	//Cancel cerebro global context cancel
	Cancel context.CancelFunc `json:"cancel" validate:"required"`

	//strategies
	strategies []strategy.Strategy `validate:"gte=1,dive,required"`

	//compress compress info map for codes
	compress []CompressInfo

	markets map[string]*market.Market

	// containers list of all container
	containers []container.Container

	//strategy.StrategyEngine embedding property for managing user strategy
	strategyEngine *strategy.Engine

	//log in cerebro global logger
	Logger Logger `validate:"required"`

	//event channel of all event
	order chan order.Order

	// eventEngine engine of management all event
	eventEngine *event.Engine

	// preload bool value, decide use candle history
	preload bool

	// dataCh all data container channel
	dataCh chan container.Container
}

//NewCerebro generate new cerebro with cerebro option
func NewCerebro(opts ...Option) *Cerebro {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Cerebro{
		Ctx:            ctx,
		Cancel:         cancel,
		compress:       []CompressInfo{},
		strategyEngine: &strategy.Engine{},
		order:          make(chan order.Order, 1),
		dataCh:         make(chan container.Container, 1),
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

//load initializing data from injected store interface
func (c *Cerebro) load() error {
	// getting live trading data like tick data
	globalch := make(chan container.Candle)

	go func() {
		for i := range globalch {
			fmt.Println(i)
		}
	}()

	c.Logger.Info("start load live data")
	if c.store == nil {
		return error2.ErrStoreNotExists
	}

	var tick <-chan container.Tick
	if err := pkg.Retry(10, func() error {
		var err error
		tick, err = c.store.LoadTick(c.Ctx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	register := make(chan string, 1)
	defer close(register)

Done:
	for {
		select {
		case tk, ok := <-tick:
			if !ok {
				break Done
			}
			if mk, ok := c.markets[tk.Code]; ok {
				select {
				case mk.Tick <- tk:
				case <-c.Ctx.Done():
					break Done
				}
			} else {
				register <- tk.Code
			}
		case code := <-register:
			mk := market.Market{
				Code: code,
				Tick: make(chan container.Tick),
			}

			go func(m *market.Market) {
				tics := []chan container.Tick{}
				for _, i := range c.compress {
					tk := make(chan container.Tick)
					tics = append(tics, tk)
					m.CompressionChans = append(m.CompressionChans, Compression(tk, i.level, i.LeftEdge))
				}

				for _, ch := range m.CompressionChans {
					go func(cha <-chan container.Candle) {
						for i := range cha {
							globalch <- i
						}
					}(ch)
				}

				go func() {
					for i := range m.Tick {
						for _, j := range tics {
							j <- i
						}
					}
				}()

			}(&mk)
			c.markets[code] = &mk
		}
	}

	return nil
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

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		c.Logger.Error(err)
		return err
	}

	c.eventEngine.Start(c.Ctx)
	c.registerEvent()
	c.broker.Store = c.store
	c.strategyEngine.Broker = c.broker
	c.strategyEngine.Start(c.Ctx, c.dataCh)

	c.orderEventRoutine()
	if err := c.load(); err != nil {
		return err
	}

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
