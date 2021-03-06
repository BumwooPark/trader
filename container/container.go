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

//container package is store of tick data or candle stick data
package container

import (
	"sync"
	"time"
)

type Container interface {
	Empty() bool
	Size() int
	Clear()
	Values() []Candle
	Add(candle Candle)
	Code() string
	Level() time.Duration
}

type SaveMode int

const (
	InMemory = iota
	External
)

type Info struct {
	Code             string
	CompressionLevel time.Duration
}

//TODO: inmemory or external storage
type DataContainer struct {
	mu         sync.RWMutex
	CandleData []Candle
	Info
}

func NewDataContainer(info Info) *DataContainer {
	return &DataContainer{
		CandleData: []Candle{},
		Info:       info,
	}
}

func (t *DataContainer) Empty() bool {
	return len(t.CandleData) == 0
}

func (t *DataContainer) Size() int {
	l := 0
	t.mu.RLock()
	l = len(t.CandleData)
	t.mu.RUnlock()
	return l
}

func (t *DataContainer) Clear() {
	t.CandleData = []Candle{}
}

func (t *DataContainer) Values() []Candle {
	t.mu.Lock()
	d := make([]Candle, len(t.CandleData))
	copy(d, t.CandleData)
	t.mu.Unlock()
	return d
}

// Add foreword append container candle data
// current candle [0] index
func (t *DataContainer) Add(candle Candle) {
	if len(t.CandleData) != 0 {
		for _, i := range t.CandleData {
			if i.Date.Equal(candle.Date) {
				return
			}
		}
	}
	t.mu.Lock()
	t.CandleData = append([]Candle{candle}, t.CandleData...)
	t.mu.Unlock()
}

func (t *DataContainer) Code() string {
	return t.Info.Code
}

func (t *DataContainer) Level() time.Duration {
	return t.Info.CompressionLevel
}
