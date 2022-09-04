package internalhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func createNewEvent() *storage.Event {
	event := &storage.Event{
		ID:          uuid.New(),
		Title:       "Занятие по Golang",
		Description: "Сервис календаря",
		UserID:      1,
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(time.Duration(5400)),
	}

	return event
}

type HandlerTestSuite struct {
	suite.Suite
	storage     *memorystorage.Storage
	logFile     *os.File
	logger      *logger.Logger
	calendar    *app.App
	ts          *httptest.Server
	updateEvent *storage.Event
	deleteEvent *storage.Event
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.storage = memorystorage.New()
	logFile, err := os.CreateTemp("", "test-logs.*.log")
	if err != nil {
		log.Fatal(err)
	}
	suite.logFile = logFile
	suite.logger = logger.New("info", suite.logFile)
	suite.calendar = app.New(suite.logger, suite.storage)
	suite.ts = httptest.NewServer(getHandler(suite.logger, suite.calendar))
	suite.updateEvent = &storage.Event{
		ID:          uuid.New(),
		Title:       "test",
		Description: "test",
		UserID:      2,
		StartDate:   time.Now().AddDate(0, 0, 1),
		EndDate:     time.Now().AddDate(0, 0, 2),
	}
	err = suite.calendar.CreateEvent(*suite.updateEvent)
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
	err = suite.calendar.CreateEvent(*suite.deleteEvent)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *HandlerTestSuite) TearDownTest() {
	os.Remove(suite.logFile.Name())
	suite.ts.Close()
}

func (suite *HandlerTestSuite) TestCreate() {
	jsonEvent, err := json.Marshal(createNewEvent())
	if err != nil {
		log.Fatal(err)
	}
	url := fmt.Sprintf("%s/create", suite.ts.URL)
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonEvent))

	response, err := http.DefaultClient.Do(request)
	suite.NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode)

	var event *storage.Event
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&event); err != nil {
		log.Fatal(err)
	}

	exists, err := suite.calendar.EventExists(event.ID)
	if err != nil {
		log.Fatal(err)
	}
	suite.True(exists)
}

func (suite *HandlerTestSuite) TestEdit() {
	jsonEvent, err := json.Marshal(struct {
		Title string
	}{
		Title: "updated",
	})
	if err != nil {
		log.Fatal(err)
	}
	url := fmt.Sprintf("%s/edit?id=%s", suite.ts.URL, suite.updateEvent.ID)
	request, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonEvent))

	response, err := http.DefaultClient.Do(request)
	suite.NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode)

	var event *storage.Event
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&event); err != nil {
		log.Fatal(err)
	}

	suite.Equal("updated", event.Title)
}

func (suite *HandlerTestSuite) TestDelete() {
	url := fmt.Sprintf("%s/delete?id=%s", suite.ts.URL, suite.deleteEvent.ID)
	request, _ := http.NewRequest(http.MethodDelete, url, nil)

	response, err := http.DefaultClient.Do(request)
	suite.NoError(err)
	suite.Equal(http.StatusNoContent, response.StatusCode)

	exists, err := suite.calendar.EventExists(suite.deleteEvent.ID)
	suite.False(exists)
}

func (suite *HandlerTestSuite) TestList() {
	date := time.Now().AddDate(0, 0, 1)
	url := fmt.Sprintf("%s/list?date=%s&duration=%s", suite.ts.URL, date.Format("2006-01-02"), storage.DayDuration)
	request, _ := http.NewRequest(http.MethodGet, url, nil)

	response, err := http.DefaultClient.Do(request)
	suite.NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode)

	responseData := make(map[uuid.UUID]storage.Event, 0)
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&responseData); err != nil {
		log.Fatal(err)
	}
	suite.Equal(1, len(responseData))
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
