package broker

import (
	"testing"
	"time"

	"github.com/gobenpark/trader/order"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestBroker_Buy(t *testing.T) {
	broker := NewBroker()
	broker.Cash = 100

	o := &order.Order{
		OType:     order.Buy,
		ExecType:  order.Limit,
		Code:      "code",
		UUID:      uuid.NewV4().String(),
		Size:      10,
		Price:     10,
		CreatedAt: time.Now(),
	}

	err := broker.Order(o)
	assert.NoError(t, err)
	assert.Equal(t, order.Submitted, o.Status())
}

func TestBroker_Cancel(t *testing.T) {
	broker := NewBroker()

	broker.Cash = 1000

	err := broker.Cancel("1123")
	assert.Error(t, err)

	t.Run("inject order", func(t *testing.T) {
		o := &order.Order{
			OType:      order.Buy,
			ExecType:   order.Limit,
			Code:       "123",
			UUID:       "123",
			Size:       10,
			Price:      1,
			CreatedAt:  time.Time{},
			ExecutedAt: time.Time{},
			StoreUID:   "",
		}
		err := broker.Order(o)
		assert.NoError(t, err)
		err = broker.Cancel("123")
		assert.NoError(t, err)
		assert.Equal(t, order.Canceled, o.Status())
	})
}
