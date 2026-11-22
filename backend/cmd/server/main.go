package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"helpdesk-backend/internal/config"
	"helpdesk-backend/internal/database"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func main() {
	cfg := config.Load()

	log.Printf("Connecting to database at %s:%s (db: %s)...", cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Printf("Warning: Database connection failed at startup: %v", err)
	} else {
		defer db.Close()
		log.Println("Database connected successfully.")

		// Jalankan migrasi dan seeding database otomatis saat startup
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("Database migration failed: %v", err)
		}
	}

	mux := http.NewServeMux()

	// Go 1.22+ ServeMux routing dengan method pattern
	mux.HandleFunc("GET /api/health", healthHandler(db))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel untuk menangkap signal shutdown (Ctrl+C atau SIGTERM dari Docker)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-shutdownChan
	log.Println("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced shutdown error: %v", err)
	}
	log.Println("Server stopped cleanly.")
}

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		dbStatus := "connected"
		isHealthy := true

		if db == nil {
			dbStatus = "disconnected"
			isHealthy = false
		} else {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err != nil {
				dbStatus = "disconnected"
				isHealthy = false
			}
		}

		res := HealthResponse{
			Status:   "ok",
			Database: dbStatus,
		}

		if !isHealthy {
			res.Status = "degraded"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
