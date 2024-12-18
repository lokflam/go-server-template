package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"server.example/internal/gke"
	"server.example/internal/httputil/graceful"
)

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func main() {
	port := "8080"

	requestLogger := httplog.NewLogger("request-logger", httplog.Options{
		JSON:          true,
		TimeFieldName: "time",
	})

	l := requestLogger.Logger

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(requestLogger))

	r.HandleFunc("GET /health", healthHandler())
	r.HandleFunc("POST /gke", gke.Handler(l))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	start := func() error {
		l.Info("starting server on " + server.Addr)
		return server.ListenAndServe()
	}

	shutdown := func(ctx context.Context) error {
		l.Info("shutting down server")
		return server.Shutdown(ctx)
	}

	err := graceful.Serve(start, shutdown, 300*time.Second)
	if err != nil {
		panic(err.Error())
	}
}
