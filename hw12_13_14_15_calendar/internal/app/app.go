package app

import (
	"context"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	Connect(ctx context.Context) error
	Close() error
	Create(s storage.Event) error
	Delete(id uuid.UUID) error
	Edit(id uuid.UUID, e storage.Event) error
	SelectForDay(date time.Time) (map[uuid.UUID]storage.Event, error)
	SelectForWeek(date time.Time) (map[uuid.UUID]storage.Event, error)
	SelectForMonth(date time.Time) (map[uuid.UUID]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(event storage.Event) {
	if err := a.storage.Create(event); err != nil {
		a.logger.Error(err.Error())
	}
}

func (a *App) DeleteEvent(id uuid.UUID) {
	if err := a.storage.Delete(id); err != nil {
		a.logger.Error(err.Error())
	}
}

func (a *App) EditEvent(id uuid.UUID, e storage.Event) {
	if err := a.storage.Edit(id, e); err != nil {
		a.logger.Error(err.Error())
	}
}

func (a *App) SelectForDay(date time.Time) map[uuid.UUID]storage.Event {
	events, err := a.storage.SelectForDay(date)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return events
}

func (a *App) SelectForWeek(date time.Time) map[uuid.UUID]storage.Event {
	events, err := a.storage.SelectForWeek(date)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return events
}

func (a *App) SelectForMonth(date time.Time) map[uuid.UUID]storage.Event {
	events, err := a.storage.SelectForMonth(date)
	if err != nil {
		a.logger.Error(err.Error())
	}

	return events
}
