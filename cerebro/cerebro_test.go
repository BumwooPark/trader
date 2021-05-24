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
	"testing"
	"time"

	"github.com/gobenpark/trader/container"
	mock_store "github.com/gobenpark/trader/store/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCerebro_loadRealtimeData(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockStore(ctrl)

	store.EXPECT().LoadTick(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (<-chan container.Tick, error) {
			ch := make(chan container.Tick)
			go func() {
				defer close(ch)
				for i := 0; i < 5; i++ {
					time.Sleep(time.Second)
					ch <- container.Tick{
						Code:   "sample",
						AskBid: "bid",
						Date:   time.Now(),
						Price:  10,
						Volume: 100,
					}
				}
			}()
			return ch, nil
		})

	c := NewCerebro(WithStore(store))

	tick, err := c.loadRealtimeData()
	assert.NoError(t, err)
	count := 0
	for range tick {
		count += 1
	}
	assert.Equal(t, 5, count)
}
