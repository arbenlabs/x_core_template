package handlers

import (
	"encoding/json"
	"net/http"
	"x/core/internal/config"
	"x/core/internal/service"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Handler struct {
	z       *zerolog.Logger
	s       *service.Service
	c       clerk.Client
	conf    config.Config
	monitor *sentryhttp.Handler
}

func NewHandler(
	logger *zerolog.Logger,
	srvc *service.Service,
	clrk clerk.Client,
	conf config.Config,
	m *sentryhttp.Handler,
) *Handler {
	return &Handler{
		z:       logger,
		s:       srvc,
		c:       clrk,
		conf:    conf,
		monitor: m,
	}
}

type APIFunc func(http.ResponseWriter, *http.Request) error

func (h *Handler) HTTPHandlerFunc(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			h.HandleErrorResponse(w, err)
		}
	}
}

func (h *Handler) WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func formatError(err error) map[string]string {
	var handlerError = err.Error()

	if err.Error() == "sql: no rows in result set" {
		handlerError = "No email found. Please sign up."
	}
	return map[string]string{
		"error": handlerError,
	}
}

func (h *Handler) HandleErrorResponse(w http.ResponseWriter, e error) error {
	return h.WriteJSON(w, http.StatusBadRequest, formatError(e))
}

func (h *Handler) RegisterRoutes() *mux.Router {
	r := mux.NewRouter()
	r.Use(h.LoggerMiddleware)
	r.Use(h.RateLimiterMiddleware)

	// sentry monitoring only for dev+prod environments
	if h.conf.Env != "local" {
		r.Use(h.monitor.Handle)
	}

	api := r.PathPrefix("/api").Subrouter()
	private := r.PathPrefix("/api").Subrouter()

	private.Use(h.ClerkAuthMiddleware)

	// Probe
	api.HandleFunc("/probe", h.HTTPHandlerFunc(h.Probe)).Methods("GET")
	return r
}
