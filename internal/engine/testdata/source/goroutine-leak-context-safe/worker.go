package worker

import (
	"context"
	"time"
)

func StartWorker(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				doWork()
				time.Sleep(time.Second)
			}
		}
	}()
}

func doWork() {}
