package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Message struct {
	Status    string    `json:"status"`
	Body      string    `json:"body"`
	Locked    bool      `json:"locked"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Handler) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		srvAddr := ctx.Value(http.LocalAddrContextKey).(net.Addr)

		slog.Info(
			"ip_address", ip,
			"host", r.URL.Host,
			"server_addr", srvAddr.String(),
			"user_agent", r.UserAgent(),
			fmt.Sprintf("%s - %s (%s)", r.Method, r.URL.Path, r.RemoteAddr),
		)

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RateLimiterMiddleware(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 5*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Lock the mutex to protect this section from race conditions.
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(3, 6)}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()

			message := Message{
				Status:    "Request Failed",
				Body:      "The account is locked or disabled. Please wait 5 minutes and try again.",
				Locked:    true,
				Timestamp: time.Now(),
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) ClerkAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type SessionName string
		var ClerkSessionName SessionName = "clerksession"

		// Get the session token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// The token should be in the format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		sessionToken := parts[1]

		// Verify the session
		session, err := h.c.VerifyToken(sessionToken)
		if err != nil {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Add the session to the request context
		ctx := context.WithValue(r.Context(), ClerkSessionName, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
