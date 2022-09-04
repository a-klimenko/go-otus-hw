package app

import (
	"context"
	"database/sql"
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
	List(date time.Time, duration string) (map[uuid.UUID]storage.Event, error)
	Exists(id uuid.UUID) (bool, error)
	GetEvent(id uuid.UUID) (storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(event storage.Event) error {
	return a.storage.Create(event)
}

func (a *App) DeleteEvent(id uuid.UUID) error {
	return a.storage.Delete(id)
}

func (a *App) EditEvent(id uuid.UUID, e storage.Event) error {
	return a.storage.Edit(id, e)
}

func (a *App) EventExists(id uuid.UUID) (bool, error) {
	return a.storage.Exists(id)
}

func (a *App) List(date time.Time, duration string) map[uuid.UUID]storage.Event {
	events, err := a.storage.List(date, duration)
	if err != nil && err != sql.ErrNoRows {
		a.logger.Error(err.Error())
	}

	return events
}

func (a *App) GetEvent(id uuid.UUID) (storage.Event, error) {
	return a.storage.GetEvent(id)
}
