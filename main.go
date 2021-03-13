package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gobenpark/trader/broker"
	"github.com/gobenpark/trader/cerebro"
	store2 "github.com/gobenpark/trader/store"
	"github.com/gobenpark/trader/strategy"
)

func main() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Println(runtime.NumGoroutine())
		}
	}()
	bk := broker.NewBroker(100000, 0.005)
	store := store2.NewStore("upbit")

	smart := &strategy.Bighands{
		Broker: bk,
	}
	smart2 := &strategy.Bighands{
		Broker: bk,
	}
	cb := cerebro.NewCerebro(
		cerebro.WithBroker(bk),
		cerebro.WithStore(store),
		cerebro.WithStrategy(smart, smart2),
		cerebro.WithResample("KRW-BTC", time.Minute),
		cerebro.WithLive(true),
		cerebro.WithPreload(true),
	)

	err := cb.Start()
	if err != nil {
		panic(err)
	}
}
