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

package broker

import (
	"errors"
	"sync"

	"github.com/gobenpark/trader/event"
	"github.com/gobenpark/trader/order"
	"github.com/gobenpark/trader/position"
	"github.com/gobenpark/trader/store"
	uuid "github.com/satori/go.uuid"
)

type Broker struct {
	sync.RWMutex
	Cash       int64
	Commission float64
	orders     map[string]*order.Order
	positions  map[string][]position.Position
	Store      store.Store
}

// NewBroker Init new broker with cash,commission
func NewBroker() *Broker {
	return &Broker{
		orders:    make(map[string]*order.Order),
		positions: make(map[string][]position.Position),
	}
}

func (b *Broker) validate(o *order.Order) error {

	if o.Size == 0 {
		return errors.New("not exist order size")
	}
	if o.Price == 0 {
		return errors.New("not exist order price")
	}
	if len(o.Code) == 0 {
		return errors.New("not exist order code")
	}

	total := float64(o.Size) * o.Price
	total = total + (total * b.Commission)

	if float64(b.Cash) < total || total == 0 {
		return errors.New("no have possible cash")
	}
	return nil
}

func (b *Broker) Order(o *order.Order) error {
	b.Lock()
	defer b.Unlock()

	if err := b.validate(o); err != nil {
		return err
	}
	if len(o.UUID) == 0 {
		o.UUID = uuid.NewV4().String()
	}
	b.orders[o.UUID] = o
	o.Submit()
	return nil
}

func (b *Broker) Cancel(uid string) error {
	o, ok := b.orders[uid]
	if !ok {
		return errors.New("not exist orderid in broker order list")
	}
	o.Cancel()
	return nil
}

func (b *Broker) Listen(e interface{}) {
	if evt, ok := e.(event.OrderEvent); ok {
		switch evt.State {
		case "cancel":
			b.Cancel(evt.Oid)
		case "done":
			//b.Accept(evt.Oid)
		case "wait":
		}
	}
}
