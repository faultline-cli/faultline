package worker

import "time"

func StartWorker() {
	go func() {
		for {
			doWork()
			time.Sleep(time.Second)
		}
	}()
}

func doWork() {}
