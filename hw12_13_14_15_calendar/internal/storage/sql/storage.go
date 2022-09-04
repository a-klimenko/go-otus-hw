package sqlstorage

import (
	"context"
	"database/sql"
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
	ctx context.Context
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

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.Open("postgres", s.dsn)
	if err != nil {
		return err
	}

	s.db = db
	s.ctx = ctx

	return nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Create(e storage.Event) error {
	query := `
				INSERT INTO events 
					(id, user_id, title, description, start_date, end_date)
				VALUES
					($1, $2, $3, $4, $5, $6)
				;
	`

	_, err := s.db.ExecContext(
		s.ctx,
		query,
		e.ID,
		e.UserID,
		e.Title,
		e.Description,
		e.StartDate,
		e.EndDate,
	)

	return err
}

func (s *Storage) Delete(id uuid.UUID) error {
	query := "DELETE FROM events WHERE id = $1"

	_, err := s.db.ExecContext(s.ctx, query, id)

	return err
}

func (s *Storage) Edit(eventID uuid.UUID, e storage.Event) error {
	query := `
				UPDATE 
					events 
				SET
					user_id = $2,
					title = $3,
					description = $4, 
					start_date = $5, 
					end_date = $6
				WHERE 
					id = $1
	`
	_, err := s.db.ExecContext(
		s.ctx,
		query,
		eventID,
		e.UserID,
		e.Title,
		e.Description,
		e.StartDate,
		e.EndDate,
	)

	return err
}

func (s *Storage) List(date time.Time, duration string) (map[uuid.UUID]storage.Event, error) {
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

func (s *Storage) SelectInDateRange(startDate time.Time, endDate time.Time) (map[uuid.UUID]storage.Event, error) {
	sql := `
		SELECT
		 id,
		 title,
		 description,
		 start_date,
		 end_date,
		 user_id
		FROM
		  events
		WHERE
		  start_date BETWEEN $1 AND $2
		;
	`
	rows, err := s.db.QueryxContext(s.ctx, sql, startDate, endDate)
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

func (s *Storage) Exists(id uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`
	err := s.db.QueryRowxContext(s.ctx, query, id).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}

func (s *Storage) GetEvent(id uuid.UUID) (storage.Event, error) {
	var event storage.Event
	query := `SELECT * FROM events WHERE id = $1`
	err := s.db.QueryRowxContext(s.ctx, query, id).StructScan(&event)

	return event, err
}
