package hw05parallelexecution

import (
	"errors"
	"sync"
)

var (
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	ErrNoWorkers           = errors.New("can't create goroutines")
)

type Task func() error

// Counter Структура для подсчета ошибок.
type Counter struct {
	mu         sync.Mutex
	errCounter int
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) GetCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.errCounter
}

func (c *Counter) IncreaseCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errCounter++
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrNoWorkers
	}

	c := NewCounter()
	taskCh := make(chan Task)
	wg := sync.WaitGroup{}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range taskCh {
				if m > 0 && m == c.GetCount() {
					return
				}

				err := task()
				if err != nil && c.GetCount() < m {
					c.IncreaseCount()
				}
			}
		}()
	}

	for _, task := range tasks {
		taskCh <- task
		if m > 0 && c.GetCount() == m {
			break
		}
	}
	close(taskCh)

	wg.Wait()

	if m > 0 && c.GetCount() == m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
