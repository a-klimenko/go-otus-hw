package memorystorage

import (
	"context"
	"fmt"
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

func (s *Storage) Connect() error {
	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Create(_ context.Context, e storage.Event) error {
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

func (s *Storage) Delete(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)

	return nil
}

func (s *Storage) Edit(_ context.Context, id uuid.UUID, e storage.Event) error {
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

func (s *Storage) List(_ context.Context, date time.Time, duration string) (map[uuid.UUID]storage.Event, error) {
	switch duration {
	case storage.DayDuration:
		return s.SelectInDateRange(date, date.AddDate(0, 0, 1))
	case storage.WeekDuration:
		return s.SelectInDateRange(date, date.AddDate(0, 0, 7))
	case storage.MonthDuration:
		return s.SelectInDateRange(date, date.AddDate(0, 1, 0))
	default:
		return s.SelectInDateRange(date, date.AddDate(0, 0, 1))
	}
}

func (s *Storage) SelectInDateRange(startDate, endDate time.Time) (map[uuid.UUID]storage.Event, error) {
	res := make(map[uuid.UUID]storage.Event, 0)

	for id, event := range s.events {
		if event.StartDate.After(startDate) && event.StartDate.Before(endDate) {
			res[id] = event
		}
	}

	return res, nil
}

func (s *Storage) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := s.events[id]

	return ok, nil
}

func (s *Storage) GetEvent(_ context.Context, id uuid.UUID) (storage.Event, error) {
	event, ok := s.events[id]

	if !ok {
		return event, fmt.Errorf("event not found")
	}

	return event, nil
}

func (s *Storage) GetByNotificationPeriod(_ context.Context, startDate, endDate time.Time) (map[uuid.UUID]storage.Event, error) {
	events, _ := s.SelectInDateRange(startDate, endDate)

	result := make(map[uuid.UUID]storage.Event, 0)
	for id, event := range events {
		if event.IsNotified == 0 {
			result[id] = event
		}
	}

	return result, nil
}

func (s *Storage) ChangeNotifyStatus(_ context.Context, eventID uuid.UUID) error {
	if event, ok := s.events[eventID]; ok {
		event.IsNotified = 1
	}

	return nil
}

func (s *Storage) DeleteOldEvents(_ context.Context) error {
	for id, event := range s.events {
		if event.EndDate.After(time.Now().AddDate(-1, 0, 0)) {
			delete(s.events, id)
		}
	}

	return nil
}
