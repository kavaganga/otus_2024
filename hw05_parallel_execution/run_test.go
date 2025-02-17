package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("tasks without errors with require.Eventually", func(t *testing.T) {
		workersCount := 5
		tasksCount := workersCount * 2 // Создаем больше задач, чем воркеров
		maxErrors := 0

		var (
			active int32                 // Счетчик одновременно выполняющихся задач
			done   = make(chan struct{}) // Канал для блокировки выполнения задач
		)

		// Создаем задачи, которые блокируются до закрытия канала done
		tasks := make([]Task, tasksCount)
		for i := 0; i < tasksCount; i++ {
			tasks[i] = func() error {
				atomic.AddInt32(&active, 1)
				defer atomic.AddInt32(&active, -1)

				<-done // Блокируем задачу до закрытия канала
				return nil
			}
		}

		// Запускаем задачи в отдельной горутине, так как Run блокирующий
		errCh := make(chan error, 1)
		go func() {
			errCh <- Run(tasks, workersCount, maxErrors)
		}()

		// Проверяем, что одновременно активно workersCount задач
		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&active) == int32(workersCount)
		}, time.Second, 10*time.Millisecond,
			"expected %d concurrent workers", workersCount)

		// Разблокируем задачи
		close(done)

		// Проверяем успешное завершение
		require.NoError(t, <-errCh, "should complete without errors")

		// Убеждаемся, что все задачи завершились
		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&active) == 0
		}, time.Second, 10*time.Millisecond,
			"all tasks should release workers")
	})
}
