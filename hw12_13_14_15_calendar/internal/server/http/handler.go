package internalhttp

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/a-klimenko/go-otus-hw/hw12_13_14_15_calendar/internal/storage"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ServerHandler struct {
	app    Application
	logger Logger
}

func (s *ServerHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	event := storage.NewEvent()
	if err := decoder.Decode(&event); err != nil {
		s.logger.Error(fmt.Sprintf("event decoding fail: %s", err))
	}

	validate := validator.New()
	if err := validate.Struct(event); err != nil {
		s.logger.Error(fmt.Sprintf("data not valid: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.app.CreateEvent(*event); err != nil {
		s.logger.Error(fmt.Sprintf("event creation fail: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(w).Encode(event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *ServerHandler) EditHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		s.logger.Error(fmt.Sprintf("can not parse id: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := s.app.GetEvent(id)
	if errors.Is(err, sql.ErrNoRows) {
		s.logger.Error(fmt.Sprintf("event not found %s", id))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		s.logger.Error(fmt.Sprintf("event decoding fail: %s", err))
	}

	validate := validator.New()
	if err := validate.Struct(event); err != nil {
		s.logger.Error(fmt.Sprintf("data not valid: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.app.EditEvent(event.ID, event); err != nil {
		s.logger.Error(fmt.Sprintf("event edition fail: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ServerHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		s.logger.Error(fmt.Sprintf("can not parse id: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, err := s.app.EventExists(id)
	if err != nil {
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exists {
		s.logger.Error(fmt.Sprintf("event not found %s", id))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := s.app.DeleteEvent(id); err != nil {
		s.logger.Error(fmt.Sprintf("event deleting fail: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *ServerHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		s.logger.Error(fmt.Sprintf("fail parse form %s", err))
	}

	date := r.URL.Query().Get("date")
	duration := r.URL.Query().Get("duration")
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		s.logger.Error(fmt.Sprintf("date parsing failed: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	list := s.app.List(parsedDate, duration)
	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getHandler(logger Logger, app Application) *http.ServeMux {
	handler := &ServerHandler{logger: logger, app: app}

	mux := http.NewServeMux()
	mux.HandleFunc("/create", handler.CreateHandler)
	mux.HandleFunc("/edit", handler.EditHandler)
	mux.HandleFunc("/delete", handler.DeleteHandler)
	mux.HandleFunc("/list", handler.ListHandler)

	return mux
}
