package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/config"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	// import postgres lib.
	_ "github.com/lib/pq"
)

type Storage struct {
	dsn string
	db  *sqlx.DB
}

func New(c config.Config) *Storage {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)

	return &Storage{
		dsn: dsn,
	}
}

func (s *Storage) Connect() error {
	db, err := sqlx.Open("postgres", s.dsn)
	if err != nil {
		return err
	}

	s.db = db

	return nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Create(ctx context.Context, e storage.Event) error {
	query := `
				INSERT INTO events 
					(id, user_id, title, description, start_date, end_date, notification_date)
				VALUES
					($1, $2, $3, $4, $5, $6, $7)
				;
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		e.ID,
		e.UserID,
		e.Title,
		e.Description,
		e.StartDate,
		e.EndDate,
		e.NotificationDate,
	)

	return err
}

func (s *Storage) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM events WHERE id = $1"

	_, err := s.db.ExecContext(ctx, query, id)

	return err
}

func (s *Storage) Edit(ctx context.Context, eventID uuid.UUID, e storage.Event) error {
	query := `
				UPDATE 
					events 
				SET
					user_id = $2,
					title = $3,
					description = $4, 
					start_date = $5, 
					end_date = $6,
					notification_date = $7
				WHERE 
					id = $1
	`
	_, err := s.db.ExecContext(
		ctx,
		query,
		eventID,
		e.UserID,
		e.Title,
		e.Description,
		e.StartDate,
		e.EndDate,
		e.NotificationDate,
	)

	return err
}

func (s *Storage) List(ctx context.Context, date time.Time, duration string) (map[uuid.UUID]storage.Event, error) {
	switch duration {
	case storage.DayDuration:
		return s.SelectInDateRange(ctx, date, date.AddDate(0, 0, 1))
	case storage.WeekDuration:
		return s.SelectInDateRange(ctx, date, date.AddDate(0, 0, 7))
	case storage.MonthDuration:
		return s.SelectInDateRange(ctx, date, date.AddDate(0, 1, 0))
	default:
		return s.SelectInDateRange(ctx, date, date.AddDate(0, 0, 1))
	}
}

func (s *Storage) SelectInDateRange(ctx context.Context, startDate time.Time, endDate time.Time) (map[uuid.UUID]storage.Event, error) {
	sql := `
		SELECT
		 id,
		 title,
		 description,
		 start_date,
		 end_date,
		 user_id,
		 notification_date
		FROM
		  events
		WHERE
		  start_date BETWEEN $1 AND $2
		;
	`
	rows, err := s.db.QueryxContext(ctx, sql, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make(map[uuid.UUID]storage.Event, 0)
	for rows.Next() {
		var event storage.Event
		err := rows.StructScan(&event)
		if err != nil {
			return nil, err
		}
		events[event.ID] = event
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Storage) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`
	err := s.db.QueryRowxContext(ctx, query, id).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return exists, nil
}

func (s *Storage) GetEvent(ctx context.Context, id uuid.UUID) (storage.Event, error) {
	var event storage.Event
	query := `SELECT * FROM events WHERE id = $1`
	err := s.db.QueryRowxContext(ctx, query, id).StructScan(&event)

	return event, err
}

func (s *Storage) GetByNotificationPeriod(ctx context.Context, startDate, endDate time.Time) (map[uuid.UUID]storage.Event, error) {
	sql := `
		SELECT
		 id,
		 title,
		 description,
		 start_date,
		 end_date,
		 user_id,
		 notification_date
		FROM
		  events
		WHERE
		  is_notified = 0
		AND
		  notification_date BETWEEN $1 AND $2
		;
	`
	rows, err := s.db.QueryxContext(ctx, sql, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make(map[uuid.UUID]storage.Event, 0)
	for rows.Next() {
		var event storage.Event
		err := rows.StructScan(&event)
		if err != nil {
			return nil, err
		}
		events[event.ID] = event
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Storage) ChangeNotifyStatus(ctx context.Context, eventID uuid.UUID) error {
	query := "UPDATE events SET is_notified = 1 WHERE id = $1"
	_, err := s.db.ExecContext(
		ctx,
		query,
		eventID,
	)

	return err
}

func (s *Storage) DeleteOldEvents(ctx context.Context) error {
	query := "DELETE FROM events WHERE end_date <= $1"

	_, err := s.db.ExecContext(
		ctx,
		query,
		time.Now().AddDate(-1, 0, 0),
	)

	return err
}
