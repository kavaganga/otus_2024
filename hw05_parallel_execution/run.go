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
	var status int32
	chTasks := make(chan Task)
	chErrors := make(chan struct{}, m)
	defer close(chErrors)

	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	atomic.StoreInt32(&status, 1)
	// Handle errors
	go func(cancel context.CancelFunc, maxErrors int, status *int32) {
		errCount := 0
		for range chErrors {
			errCount++
			if errCount >= maxErrors-1 {
				atomic.AddInt32(status, -1)
				cancel()
				return
			}
		}
	}(cancel, m, &status)

	for w := 0; w < n; w++ {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan Task, chanErrors chan<- struct{}) {
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
							chanErrors <- struct{}{}
						}
					}
				}
			}
		}(ctx, &wg, chTasks, chErrors)
	}

	for _, task := range tasks {
		if atomic.LoadInt32(&status) != 1 {
			break
		}
		chTasks <- task
	}
	close(chTasks)
	wg.Wait()

	if atomic.LoadInt32(&status) != 1 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
