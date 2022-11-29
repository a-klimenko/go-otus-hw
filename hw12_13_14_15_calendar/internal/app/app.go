package app

import (
	"context"
	"database/sql"
	"errors"
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
	Connect() error
	Close() error
	Create(ctx context.Context, s storage.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	Edit(ctx context.Context, id uuid.UUID, e storage.Event) error
	List(ctx context.Context, date time.Time, duration string) (map[uuid.UUID]storage.Event, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetEvent(ctx context.Context, id uuid.UUID) (storage.Event, error)
	GetByNotificationPeriod(ctx context.Context, startDate, endDate time.Time) (map[uuid.UUID]storage.Event, error)
	ChangeNotifyStatus(ctx context.Context, eventID uuid.UUID) error
	DeleteOldEvents(ctx context.Context) error
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.Create(ctx, event)
}

func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return a.storage.Delete(ctx, id)
}

func (a *App) EditEvent(ctx context.Context, id uuid.UUID, e storage.Event) error {
	return a.storage.Edit(ctx, id, e)
}

func (a *App) EventExists(ctx context.Context, id uuid.UUID) (bool, error) {
	return a.storage.Exists(ctx, id)
}

func (a *App) List(ctx context.Context, date time.Time, duration string) map[uuid.UUID]storage.Event {
	events, err := a.storage.List(ctx, date, duration)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		a.logger.Error(err.Error())
	}

	return events
}

func (a *App) GetEvent(ctx context.Context, id uuid.UUID) (storage.Event, error) {
	return a.storage.GetEvent(ctx, id)
}
