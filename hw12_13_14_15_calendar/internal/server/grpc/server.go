//go:generate protoc ./../../../api/EventService.proto --go_out=./eventpb --go-grpc_out=./eventpb --proto_path=./../../../
package internalgrpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/server/grpc/eventpb"
	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Server struct {
	app        Application
	logger     Logger
	server     *grpc.Server
	host, port string
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	EditEvent(ctx context.Context, id uuid.UUID, e storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, date time.Time, duration string) map[uuid.UUID]storage.Event
	GetEvent(ctx context.Context, id uuid.UUID) (storage.Event, error)
}

type CalendarService struct {
	App    Application
	Logger Logger
	eventpb.UnimplementedCalendarServer
}

func NewServer(host, port string, logger Logger, app Application) *Server {
	chainInterceptor := grpc.ChainUnaryInterceptor(
		loggingMiddleware(logger),
	)
	grpcServer := grpc.NewServer(chainInterceptor)

	service := &CalendarService{
		App:    app,
		Logger: logger,
	}
	eventpb.RegisterCalendarServer(grpcServer, service)

	srv := &Server{
		logger: logger,
		app:    app,
		server: grpcServer,
		host:   host,
		port:   port,
	}

	return srv
}

func (s *Server) Start() error {
	lsn, err := net.Listen("tcp", net.JoinHostPort(s.host, s.port))
	if err != nil {
		return err
	}

	if err := s.server.Serve(lsn); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.logger.Info("grpc server is shutting down")
	s.server.GracefulStop()

	return nil
}

func (s *CalendarService) Create(ctx context.Context, in *eventpb.CreateRequest) (*eventpb.CreateResponse, error) {
	event := storage.NewEvent()
	event.UserID = in.Event.GetUserID()
	event.Title = in.Event.GetTitle()
	event.Description = in.Event.GetDescription()
	startDate, err := time.Parse(time.RFC3339, in.Event.GetStartDate())
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can't parse start date on create event %s", err))
		return nil, err
	}
	event.StartDate = startDate
	endDate, err := time.Parse(time.RFC3339, in.Event.GetEndDate())
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can't parse end date on create event %s", err))
		return nil, err
	}
	event.EndDate = endDate
	notificationDate, err := time.Parse(time.RFC3339, in.Event.GetNotificationDate())
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can't parse notification date on create event %s", err))
		return nil, err
	}
	event.NotificationDate = notificationDate

	if err := s.App.CreateEvent(ctx, *event); err != nil {
		s.Logger.Error(fmt.Sprintf("fail create event %s", err))
		return nil, err
	}

	respEvent := eventpb.Event{}
	jsonEvent, err := json.Marshal(event)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can't marshall event on create %s", err))
		return nil, err
	}

	if err := json.Unmarshal(jsonEvent, &respEvent); err != nil {
		s.Logger.Error(fmt.Sprintf("can not create response event %s", err))
		return nil, err
	}

	return &eventpb.CreateResponse{Event: &respEvent}, nil
}

func (s *CalendarService) Edit(ctx context.Context, in *eventpb.EditRequest) (*eventpb.EditResponse, error) {
	id, err := uuid.Parse(in.GetEvent().GetID())
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can not parse id %s on event edit", id))
		return nil, err
	}

	event, err := s.App.GetEvent(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		s.Logger.Error(fmt.Sprintf("event not found %s", id))
		return nil, err
	}
	if err != nil {
		s.Logger.Error(err.Error())
		return nil, err
	}

	jsonEvent, err := json.Marshal(in.Event)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can't marshall event on edit %s", err))
		return nil, err
	}
	if err := json.Unmarshal(jsonEvent, &event); err != nil {
		s.Logger.Error(fmt.Sprintf("can not create response event %s", err))
		return nil, err
	}

	if err := s.App.EditEvent(ctx, event.ID, event); err != nil {
		s.Logger.Error(fmt.Sprintf("event editing fail: %s", err))
		return nil, err
	}

	jsonEvent, err = json.Marshal(event)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can't marshall event respone on edit %s", err))
		return nil, err
	}
	respEvent := eventpb.Event{}
	if err := json.Unmarshal(jsonEvent, &respEvent); err != nil {
		s.Logger.Error(fmt.Sprintf("can not create response event %s", err))
		return nil, err
	}

	return &eventpb.EditResponse{Event: &respEvent}, nil
}

func (s *CalendarService) Delete(ctx context.Context, in *eventpb.DeleteRequest) (*eventpb.DeleteResponse, error) {
	id, err := uuid.Parse(in.ID)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can not parse id: %s on delete", err))
		return nil, err
	}

	if err := s.App.DeleteEvent(ctx, id); err != nil {
		s.Logger.Error(fmt.Sprintf("event deleting fail: %s", err))
		return nil, err
	}

	return &eventpb.DeleteResponse{}, nil
}

func (s *CalendarService) List(ctx context.Context, in *eventpb.ListRequest) (*eventpb.ListResponse, error) {
	parsedDate, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("can't parse date on list event %s", err))
		return nil, err
	}
	list := s.App.List(ctx, parsedDate, in.Duration)

	result := make(map[string]*eventpb.Event)
	s.Logger.Info("return list events")
	for id, event := range list {
		s.Logger.Info(id.String())
		result[id.String()] = &eventpb.Event{
			ID:          event.ID.String(),
			UserID:      event.UserID,
			Title:       event.Title,
			Description: event.Description,
			StartDate:   event.StartDate.Format("2006-01-02"),
			EndDate:     event.EndDate.Format("2006-01-02"),
		}
	}

	return &eventpb.ListResponse{List: result}, nil
}
