package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/naimbob95/sharejer/internal/config"
	"github.com/naimbob95/sharejer/internal/db"
	"github.com/naimbob95/sharejer/internal/handlers"
	"github.com/naimbob95/sharejer/internal/storage"
	"github.com/naimbob95/sharejer/internal/sweeper"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	store, err := storage.NewLocalStorage(cfg.UploadsDir)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	srv := handlers.New(database, store, cfg.BaseURL, cfg.MaxUploadSize, cfg.FileExpiry)

	// ctx is cancelled when the process receives SIGINT (Ctrl+C) or SIGTERM
	// (e.g. `docker stop`). Cancelling it also stops the background sweeper.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sweeper.Start(ctx, database, store, cfg.CleanupEvery)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("GET /api/config", srv.Config)
	mux.HandleFunc("POST /api/upload", srv.FileUpload)
	mux.HandleFunc("GET /api/file/{id}", srv.FileMeta)
	mux.HandleFunc("POST /api/file/{id}/download", srv.FileDownload)
	mux.HandleFunc("DELETE /api/file/{id}", srv.FileDelete)
	mux.HandleFunc("GET /api/qr/{id}", srv.FileQR)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Run the server in its own goroutine so main can block on the shutdown
	// signal below. ListenAndServe returns ErrServerClosed on a clean shutdown,
	// which is expected — not a real error.
	go func() {
		fmt.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done() // wait for SIGINT/SIGTERM
	log.Println("shutting down…")

	// Stop accepting new connections and give in-flight requests up to 10s to
	// finish before forcing the process to exit.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("stopped")
}
