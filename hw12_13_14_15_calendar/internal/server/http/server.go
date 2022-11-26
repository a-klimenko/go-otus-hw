package internalhttp

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type Server struct {
	host, port string
	logger     Logger
	server     *http.Server
}

type EventHandler struct{}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	EditEvent(ctx context.Context, id uuid.UUID, e storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, date time.Time, duration string) map[uuid.UUID]storage.Event
	EventExists(ctx context.Context, id uuid.UUID) (bool, error)
	GetEvent(ctx context.Context, id uuid.UUID) (storage.Event, error)
}

func NewServer(host, port string, logger Logger, app Application) *Server {
	srv := &Server{
		host:   host,
		port:   port,
		logger: logger,
	}

	mux := getHandler(logger, app)
	srv.server = &http.Server{Addr: net.JoinHostPort(host, port), Handler: loggingMiddleware(mux, logger)}

	return srv
}

func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("http server is shutting down")

	return s.server.Shutdown(ctx)
}
