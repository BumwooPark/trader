package feeds

import (
	"fmt"
	"time"

	"github.com/gobenpark/trader/domain"
	"github.com/looplab/fsm"
)

const (
	HISTORYBACK = "history_back"
	LIVE        = "live"
	LOAD        = "load"
	IDLE        = "idle"
)

type DefaultFeed struct {
	Data map[string]map[time.Duration][]domain.Candle `json:"data"`
	fsm  *fsm.FSM
}

func NewFeed() *DefaultFeed {
	d := &DefaultFeed{
		Data: map[string]map[time.Duration][]domain.Candle{},
	}
	d.fsm = fsm.NewFSM(
		IDLE,
		fsm.Events{
			{Name: "historyloading", Src: []string{HISTORYBACK}, Dst: LOAD},
			{Name: "finish", Src: []string{HISTORYBACK, LOAD, IDLE, LIVE}, Dst: IDLE},
			{Name: "liveloading", Src: []string{LIVE}, Dst: LOAD},
			{Name: "readylive", Src: []string{IDLE}, Dst: LIVE},
			{Name: "loading", Src: []string{LOAD}, Dst: LOAD},
			{Name: "history", Src: []string{IDLE}, Dst: HISTORYBACK},
		},
		fsm.Callbacks{
			"enter_state": d.stateStart,
			LOAD: func(event *fsm.Event) {
				fmt.Println(event)
			},
			"finish": func(event *fsm.Event) {
				fmt.Println(event)
			},
			"liveloading":    d.loadTick,
			"historyloading": d.HistoryLoading,
		},
	)
	return d
}

func (d *DefaultFeed) HistoryLoading(e *fsm.Event) {
}

func (*DefaultFeed) stateStart(e *fsm.Event) {
	fmt.Printf("state change from %s to %s\n", e.Src, e.Dst)
}

func (d *DefaultFeed) Start(history, isLive bool) {
}

func (*DefaultFeed) Stop() {

}

func (d *DefaultFeed) Load() {

}

func (d *DefaultFeed) loadTick(e *fsm.Event) {

}

func (d *DefaultFeed) loadCandle() {

}
