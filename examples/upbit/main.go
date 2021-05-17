package main

import (
	"time"

	"github.com/gobenpark/trader/cerebro"
	"github.com/gobenpark/trader/container"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type sample struct{}

func (s *sample) Next(tick container.Tick) {
	p := message.NewPrinter(language.English)
	p.Printf("%d\n", int64(tick.Price*tick.Volume))
}

func main() {
	upbit := NewStore("upbit")
	smart := &Bighands{}

	cb := cerebro.NewCerebro(
		cerebro.WithStore(upbit, "KRW-XRP"),
		cerebro.WithStrategy(smart),
		cerebro.WithResample("KRW-XRP", time.Minute, true),
		cerebro.WithLive(true),
		cerebro.WithPreload(true),
	)

	err := cb.Start()
	if err != nil {
		panic(err)
	}
}
