package database

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	"helpdesk-backend/migrations"
)

// RunMigrations mengeksekusi semua migrasi .up.sql yang belum diterapkan ke database.
func RunMigrations(db *sql.DB) error {
	// Buat tabel schema_migrations jika belum ada
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Baca seluruh file dari embedded filesystem
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	for _, file := range upFiles {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", file).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", file, err)
		}

		if exists {
			continue // Migrasi ini sudah pernah dijalankan
		}

		log.Printf("Applying migration: %s", file)
		content, err := migrations.FS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Jalankan migrasi di dalam satu transaksi database
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}

		log.Printf("Successfully applied migration: %s", file)
	}

	return nil
}

// RollbackLastMigration membatalkan satu migrasi terakhir berdasarkan file .down.sql yang sesuai.
func RollbackLastMigration(db *sql.DB) error {
	var lastVersion string
	err := db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&lastVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No migrations to rollback.")
			return nil
		}
		return fmt.Errorf("failed to query last migration: %w", err)
	}

	downFile := strings.Replace(lastVersion, ".up.sql", ".down.sql", 1)
	log.Printf("Rolling back migration: %s", downFile)

	content, err := migrations.FS.ReadFile(downFile)
	if err != nil {
		return fmt.Errorf("down migration file not found for %s: %w", downFile, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to execute rollback %s: %w", downFile, err)
	}

	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", lastVersion); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to remove migration record %s: %w", lastVersion, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	log.Printf("Successfully rolled back migration: %s", downFile)
	return nil
}
