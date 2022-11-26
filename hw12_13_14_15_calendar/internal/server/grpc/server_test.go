package internalgrpc

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/server/grpc/eventpb"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	storage     *memorystorage.Storage
	logFile     *os.File
	logger      *logger.Logger
	calendar    *app.App
	service     *CalendarService
	updateEvent *storage.Event
	deleteEvent *storage.Event
}

func (suite *ServerTestSuite) SetupTest() {
	suite.storage = memorystorage.New()
	logFile, err := os.CreateTemp("", "test-logs.*.log")
	if err != nil {
		log.Fatal(err)
	}
	suite.logFile = logFile
	suite.logger = logger.New("info", suite.logFile)
	suite.calendar = app.New(suite.logger, suite.storage)
	suite.service = &CalendarService{
		app:    suite.calendar,
		logger: suite.logger,
	}
	suite.updateEvent = &storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      2,
		StartDate:   time.Now().AddDate(0, 0, 1),
		EndDate:     time.Now().AddDate(0, 0, 2),
	}
	err = suite.calendar.CreateEvent(context.Background(), *suite.updateEvent)
	if err != nil {
		log.Fatal(err)
	}
	suite.deleteEvent = &storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      3,
		StartDate:   time.Now().AddDate(0, 1, 0),
		EndDate:     time.Now().AddDate(0, 1, 1),
	}
	err = suite.calendar.CreateEvent(context.Background(), *suite.deleteEvent)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *ServerTestSuite) TearDownTest() {
	os.Remove(suite.logFile.Name())
}

func (suite *ServerTestSuite) TestCreate() {
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
	resp, err := suite.service.Create(context.Background(), in)
	suite.NoError(err)

	respEvent := resp.GetEvent()

	exists, err := suite.calendar.EventExists(context.Background(), uuid.MustParse(respEvent.GetID()))
	if err != nil {
		log.Fatal(err)
	}
	suite.True(exists)
}

func (suite *ServerTestSuite) TestEdit() {
	event := &eventpb.Event{
		ID:    suite.updateEvent.ID.String(),
		Title: "updated",
	}

	in := &eventpb.EditRequest{
		Event: event,
	}
	resp, err := suite.service.Edit(context.Background(), in)
	suite.NoError(err)

	respEvent := resp.GetEvent()

	suite.Equal("updated", respEvent.GetTitle())
}

func (suite *ServerTestSuite) TestDelete() {
	in := &eventpb.DeleteRequest{
		ID: suite.deleteEvent.ID.String(),
	}
	_, err := suite.service.Delete(context.Background(), in)
	suite.NoError(err)

	exists, err := suite.calendar.EventExists(context.Background(), suite.deleteEvent.ID)
	suite.NoError(err)
	suite.False(exists)
}

func (suite *ServerTestSuite) TestList() {
	date := time.Now().AddDate(0, 0, 1)
	in := &eventpb.ListRequest{
		Date:     date.Format("2006-01-02"),
		Duration: storage.DayDuration,
	}

	resp, err := suite.service.List(context.Background(), in)
	suite.NoError(err)
	suite.Equal(1, len(resp.List))
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
