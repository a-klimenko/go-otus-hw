package memorystorage

import (
	"errors"
	"testing"
	"time"

	memorystorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	storage := New()
	movieEvent := memorystorage.Event{
		ID:          uuid.New(),
		Title:       "Watching a movie",
		Description: "Watching the new blockbuster with friends",
		UserID:      1,
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(time.Hour * 3),
	}
	eventID := movieEvent.ID
	err := storage.Create(movieEvent)
	require.NoError(t, err)

	t.Run("busy error", func(t *testing.T) {
		require.Len(t, storage.events, 1)

		lessonEvent := memorystorage.Event{
			ID:          uuid.New(),
			Title:       "Go course lesson",
			Description: "New lesson in go developer course",
			UserID:      1,
			StartDate:   time.Now().Add(time.Hour),
			EndDate:     time.Now().Add(time.Hour * 2),
		}
		err := storage.Create(lessonEvent)

		require.True(t, errors.Is(err, memorystorage.ErrDateAlreadyBusy))
	})

	t.Run("edit", func(t *testing.T) {
		movieEvent.Title = "Watch the new movie"

		eventsForDayBeforeEdit, err := storage.List(movieEvent.StartDate.Add(-1*time.Hour), memorystorage.DayDuration)
		require.NoError(t, err)
		require.Len(t, eventsForDayBeforeEdit, 1)

		firstDayEvent := eventsForDayBeforeEdit[eventID]
		require.NotEqual(t, firstDayEvent, movieEvent)

		err = storage.Edit(movieEvent.ID, movieEvent)
		require.NoError(t, err)

		eventsForDayAfterEdit, err := storage.List(movieEvent.StartDate.Add(-1*time.Hour), memorystorage.DayDuration)
		require.NoError(t, err)

		updatedEvent := eventsForDayAfterEdit[eventID]
		require.Equal(t, movieEvent, updatedEvent)
	})

	t.Run("select", func(t *testing.T) {
		emptyForDayEvents, err := storage.List(
			movieEvent.StartDate.AddDate(0, 0, -2),
			memorystorage.DayDuration,
		)
		require.NoError(t, err)
		require.Len(t, emptyForDayEvents, 0)

		forDayEvents, err := storage.List(
			movieEvent.StartDate.Add(-1*time.Hour),
			memorystorage.DayDuration,
		)
		require.NoError(t, err)
		require.Len(t, forDayEvents, 1)
		require.Equal(t, movieEvent, forDayEvents[eventID])

		emptyForWeekEvents, err := storage.List(
			movieEvent.StartDate.AddDate(0, 0, -8),
			memorystorage.WeekDuration,
		)
		require.NoError(t, err)
		require.Len(t, emptyForWeekEvents, 0)

		forWeekEvents, err := storage.List(
			movieEvent.StartDate.AddDate(0, 0, -1),
			memorystorage.WeekDuration,
		)
		require.NoError(t, err)
		require.Len(t, forWeekEvents, 1)
		require.Equal(t, movieEvent, forWeekEvents[eventID])

		emptyForMonths, err := storage.List(
			movieEvent.StartDate.AddDate(0, -1, -1),
			memorystorage.MonthDuration,
		)
		require.NoError(t, err)
		require.Len(t, emptyForMonths, 0)

		forMonths, err := storage.List(
			movieEvent.StartDate.AddDate(0, 0, -1),
			memorystorage.MonthDuration,
		)
		require.NoError(t, err)
		require.Len(t, forMonths, 1)
		require.Equal(t, movieEvent, forMonths[eventID])
	})

	t.Run("delete", func(t *testing.T) {
		err = storage.Delete(movieEvent.ID)
		require.NoError(t, err)

		eventsForDay, err := storage.List(
			movieEvent.StartDate.Add(-1*time.Hour),
			memorystorage.MonthDuration,
		)
		require.NoError(t, err)
		require.Len(t, eventsForDay, 0)

		err = storage.Delete(uuid.New())
		require.True(t, errors.Is(err, memorystorage.ErrEventNotFound))
	})
}
