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
package container

import "time"

type Candle struct {
	Code   string    `json:"code" validate:"required"`
	Open   float64   `json:"open" validate:"required"`
	High   float64   `json:"high" validate:"required"`
	Low    float64   `json:"low" validate:"required"`
	Close  float64   `json:"close" validate:"required"`
	Volume float64   `json:"volume" validate:"required"`
	Date   time.Time `json:"date" validate:"required"`
}
