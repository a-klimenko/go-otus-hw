package integration_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/config"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/rabbitmq"
	internalscheduler "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/scheduler"
	internalsender "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/sender"
	sqlstorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage/sql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/server/grpc/eventpb"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type CalendarTestSuite struct {
	suite.Suite
	config   config.Config
	storage  *sqlstorage.Storage
	logFile  *os.File
	logger   *logger.Logger
	calendar *app.App
	client   eventpb.CalendarClient
}

func (suite *CalendarTestSuite) SetupTest() {
	suite.config = config.Config{
		DBUser:           "user",
		DBPassword:       "1234",
		DBHost:           "db",
		DBPort:           "5432",
		DBName:           "calendar",
		AMQPAddress:      "amqp://guest:guest@rabbitmq:5672",
		AMQPQueueName:    "event_notifications",
		EventScanTimeout: 60 * time.Second,
	}
	suite.storage = sqlstorage.New(suite.config)
	suite.storage.Connect()
	logFile, err := os.CreateTemp("", "test-logs.*.log")
	if err != nil {
		log.Fatal(err)
	}
	suite.logFile = logFile
	suite.logger = logger.New("info", suite.logFile)
	suite.calendar = app.New(suite.logger, suite.storage)
	host := net.JoinHostPort("calendar", "50051")
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	suite.client = eventpb.NewCalendarClient(conn)
}

func (suite *CalendarTestSuite) TearDownTest() {
	os.Remove(suite.logFile.Name())
	suite.storage.Close()
}

func (suite *CalendarTestSuite) TestCreate() {
	suite.Run("create success", func() {
		event := &eventpb.Event{
			Title:            "Занятие по Golang",
			Description:      "Сервис календаря",
			UserID:           1,
			StartDate:        time.Now().Format(time.RFC3339),
			EndDate:          time.Now().Add(time.Duration(5400)).Format(time.RFC3339),
			NotificationDate: time.Now().Add(-time.Duration(400)).Format(time.RFC3339),
		}
		in := &eventpb.CreateRequest{
			Event: event,
		}

		resp, err := suite.client.Create(context.Background(), in)
		suite.NoError(err)

		respEvent := resp.GetEvent()

		exists, err := suite.calendar.EventExists(context.Background(), uuid.MustParse(respEvent.GetID()))
		suite.NoError(err)
		suite.True(exists)
		suite.NoError(suite.storage.Delete(context.Background(), uuid.MustParse(respEvent.GetID())))
	})

	suite.Run("create date busy", func() {
		event := &eventpb.Event{
			Title:            "test",
			Description:      "test",
			UserID:           2,
			StartDate:        time.Now().Add(time.Duration(1000)).Format(time.RFC3339),
			EndDate:          time.Now().Add(time.Duration(5400)).Format(time.RFC3339),
			NotificationDate: time.Now().Format(time.RFC3339),
		}
		in := &eventpb.CreateRequest{
			Event: event,
		}

		resp, err := suite.client.Create(context.Background(), in)
		suite.NoError(err)

		_, err = suite.client.Create(context.Background(), in)
		suite.Error(err)
		suite.NoError(suite.storage.Delete(context.Background(), uuid.MustParse(resp.GetEvent().GetID())))
	})
}

func (suite *CalendarTestSuite) TestEdit() {
	event := storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      2,
		StartDate:   time.Now().AddDate(0, 0, 1),
		EndDate:     time.Now().AddDate(0, 0, 2),
	}
	suite.NoError(suite.calendar.CreateEvent(context.Background(), event))

	pbEvent := &eventpb.Event{
		ID:    event.ID.String(),
		Title: "updated",
	}

	in := &eventpb.EditRequest{
		Event: pbEvent,
	}
	resp, err := suite.client.Edit(context.Background(), in)
	suite.NoError(err)

	respEvent := resp.GetEvent()
	editedEvent, err := suite.storage.GetEvent(context.Background(), event.ID)
	suite.NoError(err)

	suite.Equal("updated", editedEvent.Title)
	suite.Equal(respEvent.GetID(), editedEvent.ID.String())
	suite.NoError(suite.storage.Delete(context.Background(), editedEvent.ID))
}

func (suite *CalendarTestSuite) TestDelete() {
	deleteEvent := storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      3,
		StartDate:   time.Now().AddDate(0, 1, 0),
		EndDate:     time.Now().AddDate(0, 1, 1),
	}
	suite.NoError(suite.calendar.CreateEvent(context.Background(), deleteEvent))

	in := &eventpb.DeleteRequest{
		ID: deleteEvent.ID.String(),
	}
	_, err := suite.client.Delete(context.Background(), in)
	suite.NoError(err)

	exists, err := suite.calendar.EventExists(context.Background(), deleteEvent.ID)
	suite.NoError(err)
	suite.False(exists)
}

func (suite *CalendarTestSuite) TestList() {
	event1 := storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      3,
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 1),
	}
	suite.NoError(suite.calendar.CreateEvent(context.Background(), event1))

	event2 := storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      3,
		StartDate:   time.Now().AddDate(0, 0, 6),
		EndDate:     time.Now().AddDate(0, 0, 7),
	}
	suite.NoError(suite.calendar.CreateEvent(context.Background(), event2))

	event3 := storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      3,
		StartDate:   time.Now().AddDate(0, 1, -1),
		EndDate:     time.Now().AddDate(0, 1, 0),
	}
	suite.NoError(suite.calendar.CreateEvent(context.Background(), event3))

	suite.Run("list for day", func() {
		date := time.Now()
		in := &eventpb.ListRequest{
			Date:     date.Format("2006-01-02"),
			Duration: storage.DayDuration,
		}

		resp, err := suite.client.List(context.Background(), in)
		suite.NoError(err)
		suite.Equal(1, len(resp.List))
	})

	suite.Run("list for week", func() {
		date := time.Now()
		in := &eventpb.ListRequest{
			Date:     date.Format("2006-01-02"),
			Duration: storage.WeekDuration,
		}

		resp, err := suite.client.List(context.Background(), in)
		suite.NoError(err)
		suite.Equal(2, len(resp.List))
	})

	suite.Run("list for month", func() {
		date := time.Now()
		in := &eventpb.ListRequest{
			Date:     date.Format("2006-01-02"),
			Duration: storage.MonthDuration,
		}

		resp, err := suite.client.List(context.Background(), in)
		suite.NoError(err)
		suite.Equal(3, len(resp.List))
	})

	suite.NoError(suite.storage.Delete(context.Background(), event1.ID))
	suite.NoError(suite.storage.Delete(context.Background(), event2.ID))
	suite.NoError(suite.storage.Delete(context.Background(), event3.ID))
}

func (suite *CalendarTestSuite) TestSend() {
	event := storage.Event{
		ID:               uuid.New(),
		Title:            "test",
		Description:      "test",
		UserID:           3,
		StartDate:        time.Now(),
		EndDate:          time.Now().AddDate(0, 0, 1),
		NotificationDate: time.Now(),
	}
	suite.NoError(suite.calendar.CreateEvent(context.Background(), event))
	rabbit := rabbitmq.New(suite.config)
	err := rabbit.Connect()
	suite.NoError(err)

	scheduler := internalscheduler.New(suite.config, suite.storage, suite.logger, rabbit)
	sender := internalsender.New(suite.storage, suite.logger, rabbit)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := scheduler.Run(ctx); err != nil {
			suite.logger.Info(fmt.Sprintf("failed to start scheduler: %s", err))
		}
		wg.Done()
	}()
	wg.Wait()
	wg.Add(1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := sender.Run(ctx); err != nil {
			suite.logger.Info(fmt.Sprintf("failed to start sender: %s", err))
		}
		wg.Done()
	}()
	wg.Wait()
	scheduler.Shutdown()
	sender.Shutdown()
	notifiedEvent, err := suite.storage.GetEvent(context.Background(), event.ID)
	suite.NoError(err)
	suite.True(notifiedEvent.IsNotified == 1)
	suite.NoError(suite.storage.Delete(context.Background(), event.ID))
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(CalendarTestSuite))
}
