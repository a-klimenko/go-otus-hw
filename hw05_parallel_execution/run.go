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

type Counter struct {
	blocker    chan struct{}
	taskCh     chan Task
	quitCh     chan struct{}
	factoryWg  sync.WaitGroup
	workersWg  sync.WaitGroup
	tasksWg    sync.WaitGroup
	mu         sync.Mutex
	errCounter int
}

func NewCounter(n int) *Counter {
	return &Counter{
		blocker: make(chan struct{}, n),
		taskCh:  make(chan Task),
		quitCh:  make(chan struct{}),
	}
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

	c := NewCounter(n)

	c.factoryWg.Add(n)
	go RunFactory(m, c)
	c.factoryWg.Wait()

	for _, task := range tasks {
		c.taskCh <- task
		c.tasksWg.Add(1)
		if m > 0 && c.GetCount() == m {
			break
		}
	}
	close(c.taskCh)

	c.tasksWg.Wait()
	c.quitCh <- struct{}{}
	close(c.quitCh)
	c.workersWg.Wait()

	if m > 0 && c.GetCount() == m {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func RunFactory(m int, c *Counter) {
	defer close(c.blocker)
	for {
		select {
		case <-c.quitCh:
			return
		case c.blocker <- struct{}{}:
			c.factoryWg.Done()
			c.workersWg.Add(1)
			go func() {
				task, ok := <-c.taskCh
				defer c.workersWg.Done()
				if ok {
					c.factoryWg.Add(1)
					defer func() { <-c.blocker }()
				}
				if task != nil {
					defer c.tasksWg.Done()
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
	}
}
