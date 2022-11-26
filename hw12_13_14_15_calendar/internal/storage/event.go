package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	DayDuration   = "day"
	WeekDuration  = "week"
	MonthDuration = "month"
)

var (
	ErrDateAlreadyBusy = errors.New("date already busy")
	ErrEventNotFound   = errors.New("event not exists")
)

type Event struct {
	ID          uuid.UUID `db:"id"`
	Title       string    `db:"title" json:"title" validate:"required"`
	Description string    `db:"description" json:"description"`
	UserID      int64     `db:"user_id" json:"userId" validate:"numeric,required"`
	StartDate   time.Time `db:"start_date" json:"startDate" validate:"required"`
	EndDate     time.Time `db:"end_date" json:"endDate" validate:"required"`
}

func NewEvent() *Event {
	event := &Event{
		ID: uuid.New(),
	}

	return event
}
