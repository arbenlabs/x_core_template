package handlers

import (
	"net/http"
	"time"
)

type ProbeResponse struct {
	Status      ServerStatus `json:"status"`
	Message     string       `json:"message"`
	Environment string       `json:"environment"`
	DateTime    time.Time    `json:"timestamp"`
}

type ServerStatus string

const (
	ServerStatusHealthy ServerStatus = "HEALTHY"
	ServerStatusError   ServerStatus = "ERROR"
)

func (h *Handler) Probe(w http.ResponseWriter, r *http.Request) error {
	h.z.Info().Msg("probing backend server")

	pong := ProbeResponse{
		Status:      ServerStatusHealthy,
		Message:     "pong",
		DateTime:    time.Now(),
		Environment: h.conf.Env,
	}

	return h.WriteJSON(w, http.StatusOK, pong)
}
