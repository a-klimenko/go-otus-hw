package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
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
	CreateEvent(event storage.Event) error
	EditEvent(id uuid.UUID, e storage.Event) error
	DeleteEvent(id uuid.UUID) error
	List(date time.Time, duration string) map[uuid.UUID]storage.Event
	EventExists(id uuid.UUID) (bool, error)
	GetEvent(id uuid.UUID) (storage.Event, error)
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

func (s *Server) Start(ctx context.Context) error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error(fmt.Sprintf("listen: %s", err))
		os.Exit(1)
	}
	<-ctx.Done()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("calendar is shutting down")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error(fmt.Sprintf("shutdown: %s", err))
	}
	<-ctx.Done()

	return nil
}
