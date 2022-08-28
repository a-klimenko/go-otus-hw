package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type Storage struct {
	events map[uuid.UUID]storage.Event
	mu     sync.RWMutex
}

func New() *Storage {
	return &Storage{
		events: make(map[uuid.UUID]storage.Event),
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Create(e storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range s.events {
		if event.StartDate.Before(e.StartDate) && event.EndDate.After(e.StartDate) {
			return storage.ErrDateAlreadyBusy
		}
	}

	s.events[e.ID] = e

	return nil
}

func (s *Storage) Delete(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)

	return nil
}

func (s *Storage) Edit(id uuid.UUID, e storage.Event) error {
	s.mu.RLock()
	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}
	s.mu.RUnlock()

	s.mu.Lock()
	s.events[id] = e
	s.mu.Unlock()

	return nil
}

func (s *Storage) SelectForDay(date time.Time) (map[uuid.UUID]storage.Event, error) {
	return s.list(date, date.AddDate(0, 0, 1))
}

func (s *Storage) SelectForWeek(date time.Time) (map[uuid.UUID]storage.Event, error) {
	return s.list(date, date.AddDate(0, 0, 7))
}

func (s *Storage) SelectForMonth(date time.Time) (map[uuid.UUID]storage.Event, error) {
	return s.list(date, date.AddDate(0, 1, 0))
}

func (s *Storage) list(startDate, endDate time.Time) (map[uuid.UUID]storage.Event, error) {
	res := make(map[uuid.UUID]storage.Event, 0)

	for id, event := range s.events {
		if event.StartDate.After(startDate) && event.StartDate.Before(endDate) {
			res[id] = event
		}
	}

	return res, nil
}
