package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	var countErrors int32
	chTasks := make(chan Task)
	atomic.StoreInt32(&countErrors, int32(m))
	wg := sync.WaitGroup{}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, tasks <-chan Task) {
			defer wg.Done()
			for task := range tasks {
				if atomic.LoadInt32(&countErrors) < 0 {
					return
				}
				if err := task(); err != nil {
					atomic.AddInt32(&countErrors, -1)
				}
			}
		}(&wg, chTasks)
	}

	for _, task := range tasks {
		if atomic.LoadInt32(&countErrors) < 0 {
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
