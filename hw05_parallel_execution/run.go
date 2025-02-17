package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	chTasks := make(chan Task)
	chErrors := make(chan struct{}, m)
	defer close(chErrors)

	wg := sync.WaitGroup{}
	status := true

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

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

	// Handle errors
	go func(cancel context.CancelFunc, maxErrors int, status *bool) {
		errCount := 0
		for range chErrors {
			errCount++
			if errCount >= maxErrors-1 {
				*status = false
				cancel()
				return
			}
		}
	}(cancel, m, &status)

	for _, task := range tasks {
		if !status {
			break
		}
		chTasks <- task
	}
	close(chTasks)

	wg.Wait()

	if !status {
		return ErrErrorsLimitExceeded
	}
	return nil
}
