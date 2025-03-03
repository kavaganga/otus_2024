package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	var countErrors int32
	chTasks := make(chan Task, len(tasks))
	atomic.StoreInt32(&countErrors, int32(m))
	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for w := 0; w < n; w++ {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan Task, cancel context.CancelFunc) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-tasks:
					if !ok {
						return
					} else {
						if err := task(); err != nil {
							atomic.AddInt32(&countErrors, -1)
							if atomic.LoadInt32(&countErrors) < 0 {
								cancel()
								return
							}
						}
					}
				}
			}
		}(ctx, &wg, chTasks, cancel)
	}

	for _, task := range tasks {
		if atomic.LoadInt32(&countErrors) < 0 {
			cancel()
			break
		}
		chTasks <- task
	}
	close(chTasks)
	wg.Wait()
	if countErrors < 0 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
