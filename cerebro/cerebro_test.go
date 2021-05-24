/*
 *  Copyright 2021 The Trader Authors
 *
 *  Licensed under the GNU General Public License v3.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      <https:fsf.org/>
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package cerebro

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/gobenpark/trader/container"
	"github.com/gobenpark/trader/event"
	"github.com/gobenpark/trader/order"
	"github.com/gobenpark/trader/position"
	"github.com/stretchr/testify/assert"
)

type SampleStore struct {
}

func (s SampleStore) Order(o *order.Order) error {
	panic("implement me")
}

func (s SampleStore) Cancel(id string) error {
	panic("implement me")
}

func (s SampleStore) LoadHistory(ctx context.Context, d time.Duration) ([]container.Candle, error) {
	return []container.Candle{}, nil
}

func (s SampleStore) LoadTick(ctx context.Context) (<-chan container.Tick, error) {
	ch := make(chan container.Tick)
	data := []string{"code1", "code2", "code3"}
	go func() {
		defer close(ch)
		for {
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(10)))
			ch <- container.Tick{
				Code:   data[rand.Intn(3)],
				AskBid: "ASK",
				Date:   time.Now(),
				Price:  float64(rand.Intn(1000000)),
				Volume: float64(rand.Int63n(10000000)),
			}
		}
	}()
	return ch, nil
}

func (s SampleStore) Uid() string {
	panic("implement me")
}

func (s SampleStore) Cash() int64 {
	panic("implement me")
}

func (s SampleStore) Commission() float64 {
	panic("implement me")
}

func (s SampleStore) Positions() []position.Position {
	panic("implement me")
}

func (s SampleStore) OrderState(ctx context.Context) (<-chan event.OrderEvent, error) {
	panic("implement me")
}

func (s SampleStore) OrderInfo(id string) (*order.Order, error) {
	panic("implement me")
}

func TestNewCerebro(t *testing.T) {
	tests := []struct {
		name    string
		cerebro *Cerebro
		checker func(c *Cerebro, t *testing.T)
	}{
		{
			"default broker",
			NewCerebro(),
			func(c *Cerebro, t *testing.T) {
				assert.NotNil(t, c.broker)
			},
		},
		{
			"preload false",
			NewCerebro(),
			func(c *Cerebro, t *testing.T) {
				assert.False(t, c.preload)
			},
		},
		{
			"preload true",
			NewCerebro(WithPreload(true)),
			func(c *Cerebro, t *testing.T) {
				assert.True(t, c.preload)
			},
		},
		{
			"cerebro order channel exist",
			NewCerebro(),
			func(c *Cerebro, t *testing.T) {
				assert.NotNil(t, c.order)
			},
		},
		{
			"cerebro data container not exist",
			NewCerebro(),
			func(c *Cerebro, t *testing.T) {
				assert.Nil(t, c.containers)
			},
		},
		{
			"cerebro strategy engine exist",
			NewCerebro(),
			func(c *Cerebro, t *testing.T) {
				assert.NotNil(t, c.strategyEngine)
			},
		},
		{
			"container not exist",
			NewCerebro(),
			func(c *Cerebro, t *testing.T) {
				assert.Nil(t, c.getContainer("nil", time.Second*0))
				assert.Nil(t, c.containers)
			},
		},
		{
			"marketProcess",
			func() *Cerebro {
				s := SampleStore{}

				c := NewCerebro(
					WithStore(s),
					WithResample([]time.Duration{3 * time.Minute, 1 * time.Minute, 3 * time.Second}, true),
				)
				return c
			}(),
			func(c *Cerebro, t *testing.T) {
				t.Parallel()

				c.load()

				time.Sleep(2 * time.Second)

			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.checker(test.cerebro, t)
		})
	}
}

func TestCerebro_Broker_Exist(t *testing.T) {
	NewCerebro()
}
