//go:generate protoc ./../../../api/EventService.proto --go_out=./eventpb --go-grpc_out=./eventpb --proto_path=./../../../
package internalgrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
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
	CreateEvent(event storage.Event) error
	EditEvent(id uuid.UUID, e storage.Event) error
	DeleteEvent(id uuid.UUID) error
	List(date time.Time, duration string) map[uuid.UUID]storage.Event
}

type CalendarService struct {
	app    Application
	logger Logger
	eventpb.UnimplementedCalendarServer
}

func NewServer(host, port string, logger Logger, app Application) *Server {
	chainInterceptor := grpc.ChainUnaryInterceptor(
		loggingMiddleware(logger),
	)
	grpcServer := grpc.NewServer(chainInterceptor)

	service := &CalendarService{
		app:    app,
		logger: logger,
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

func (s *Server) Start(ctx context.Context) error {
	lsn, err := net.Listen("tcp", net.JoinHostPort(s.host, s.port))
	if err != nil {
		s.logger.Error(fmt.Sprintf("fail start gprc server: %s", err))
	}

	if err := s.server.Serve(lsn); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error(fmt.Sprintf("listen: %s", err))
		os.Exit(1)
	}
	<-ctx.Done()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("calendar is shutting down")
	s.server.GracefulStop()
	<-ctx.Done()

	return nil
}

func (s *CalendarService) Create(_ context.Context, in *eventpb.CreateRequest) (*eventpb.CreateResponse, error) {
	encodedEvent, _ := json.Marshal(in.Event)
	event := storage.NewEvent()
	event.UserID = in.Event.GetUserID()
	event.Title = in.Event.GetTitle()
	event.Description = in.Event.GetDescription()
	event.StartDate, _ = time.Parse(time.RFC3339, in.Event.GetStartDate())
	event.EndDate, _ = time.Parse(time.RFC3339, in.Event.GetEndDate())
	if err := json.Unmarshal(encodedEvent, event); err != nil {
		return nil, err
	}

	if err := s.app.CreateEvent(*event); err != nil {
		s.logger.Error(fmt.Sprintf("fail create event %s", err))
		return nil, err
	}

	return &eventpb.CreateResponse{}, nil
}

func (s *CalendarService) Edit(_ context.Context, in *eventpb.EditRequest) (*eventpb.EditResponse, error) {
	encodedEvent, _ := json.Marshal(in.Event)
	var event storage.Event
	if err := json.Unmarshal(encodedEvent, &event); err != nil {
		return nil, err
	}

	if err := s.app.EditEvent(event.ID, event); err != nil {
		s.logger.Error(fmt.Sprintf("event editing fail: %s", err))
		return nil, err
	}

	return &eventpb.EditResponse{}, nil
}

func (s *CalendarService) Delete(_ context.Context, in *eventpb.DeleteRequest) (*eventpb.DeleteResponse, error) {
	if err := s.app.DeleteEvent(uuid.MustParse(in.ID)); err != nil {
		s.logger.Error(fmt.Sprintf("event deleting fail: %s", err))
		return nil, err
	}

	return &eventpb.DeleteResponse{}, nil
}

func (s *CalendarService) List(_ context.Context, in *eventpb.ListRequest) (*eventpb.ListResponse, error) {
	parsedDate, _ := time.Parse("2006-01-02", in.Date)
	list := s.app.List(parsedDate, in.Duration)

	result := make(map[string]*eventpb.Event)
	for id, event := range list {
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
