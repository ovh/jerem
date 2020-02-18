package core

import (
	"sync"
	"time"
)

// Runner strunct handling runners internal
type Runner struct {
	ticker *time.Ticker
	runner func()
	wg     sync.WaitGroup
	end    chan *struct{}
}

// NewRunner start a new runner
func NewRunner(runner func(), interval time.Duration) *Runner {
	r := Runner{
		ticker: time.NewTicker(interval),
		runner: runner,
		end:    make(chan *struct{}),
	}

	go r.run()

	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.run()
			case <-r.end:
				return
			}
		}
	}()

	return &r
}

func (r *Runner) run() {
	r.wg.Add(1)
	r.runner()
	r.wg.Done()
}

// Stop stop the runner
func (r *Runner) Stop() {
	r.ticker.Stop()
	r.end <- nil
	r.wg.Wait()
}
