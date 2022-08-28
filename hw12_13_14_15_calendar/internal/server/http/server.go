package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type Server struct {
	host, port string
	app        Application
	logger     Logger
	server     *http.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Application interface {
	CreateEvent(event storage.Event)
	EditEvent(id uuid.UUID, e storage.Event)
	DeleteEvent(id uuid.UUID)
	SelectForDay(date time.Time) map[uuid.UUID]storage.Event
	SelectForWeek(date time.Time) map[uuid.UUID]storage.Event
	SelectForMonth(date time.Time) map[uuid.UUID]storage.Event
}

func NewServer(host, port string, logger Logger, app Application) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello-world"))
	})
	srv := &http.Server{Addr: host + ":" + port, Handler: loggingMiddleware(mux, logger)}

	return &Server{
		host:   host,
		port:   port,
		app:    app,
		logger: logger,
		server: srv,
	}
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
