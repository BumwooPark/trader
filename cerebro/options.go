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
	"time"

	"github.com/gobenpark/trader/store"
	"github.com/gobenpark/trader/strategy"
)

type Option func(*Cerebro)

func WithCash(cash int64) Option {
	return func(c *Cerebro) {
		c.broker.Cash = cash
	}
}

func WithCommission(commission float64) Option {
	return func(c *Cerebro) {
		c.broker.Commission = commission
	}
}

//func WithObserver(o observer.Observer) Option {
//	return func(c *Cerebro) {
//		c.o = o
//	}
//}

func WithStrategy(s ...strategy.Strategy) Option {
	return func(c *Cerebro) {
		c.strategyEngine.Sts = s
	}
}

func WithStore(s store.Store) Option {
	return func(c *Cerebro) {
		c.store = s
	}
}
func WithResample(level []time.Duration, leftEdge bool) Option {
	return func(c *Cerebro) {
		for _, i := range level {
			c.compress = append(c.compress, CompressInfo{level: i, LeftEdge: leftEdge})
		}
	}
}
