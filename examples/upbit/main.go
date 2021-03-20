package main

import (
	"time"

	"github.com/gobenpark/trader/broker"
	"github.com/gobenpark/trader/cerebro"
	store2 "github.com/gobenpark/trader/store"
	"github.com/gobenpark/trader/strategy"
)

func main() {
	//go func() {
	//	for {
	//		time.Sleep(1 * time.Second)
	//		fmt.Println(runtime.NumGoroutine())
	//	}
	//}()
	bk := broker.NewBroker(100000, 0.005)

	wax := store2.NewStore("upbit", "KRW-WAXP")
	mlk := store2.NewStore("KRW-MLK", "KRW-MLK")
	ong := store2.NewStore("upbit", "KRW-ONG")
	bora := store2.NewStore("bora", "KRW-BORA")
	lbc := store2.NewStore("lbc", "KRW-LBC")

	smart := &strategy.Bighands{
		Broker: bk,
	}

	cb := cerebro.NewCerebro(
		cerebro.WithBroker(bk),
		cerebro.WithStore(wax, mlk, ong, bora, lbc),
		cerebro.WithStrategy(smart),
		cerebro.WithResample(wax, time.Minute*3, true),
		cerebro.WithResample(mlk, time.Minute*3, true),
		cerebro.WithResample(ong, time.Minute*3, true),
		cerebro.WithResample(bora, time.Minute*3, true),
		cerebro.WithResample(lbc, time.Minute*3, true),
		cerebro.WithLive(true),
		cerebro.WithPreload(true),
	)

	err := cb.Start()
	if err != nil {
		panic(err)
	}
}