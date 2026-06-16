package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"back-retro-log/internal/app"
	"back-retro-log/internal/auth"
	"back-retro-log/internal/config"
	"back-retro-log/internal/db"
	"back-retro-log/internal/providers"

	_ "modernc.org/sqlite"
)

//go:embed sql/migrations/*.sql
var migrationFS embed.FS

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	sqlDB, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(1)

	if err := runMigrations(sqlDB); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	queries := db.New(sqlDB)
	sessionMgr := auth.NewSessionManager(queries)

	var gameProvider providers.GameProvider
	if cfg.RAWGKey != "" {
		gameProvider = providers.NewRAWG(cfg.RAWGKey)
	} else {
		gameProvider = providers.NewDummy()
		log.Println("WARNING: RAWG_API_KEY not set, using dummy provider")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	router := app.NewRouter(queries, sessionMgr, gameProvider, baseURL)

	go func() {
		sessionMgr.Cleanup()
	}()

	log.Printf("Starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runMigrations(sqlDB *sql.DB) error {
	entries, err := migrationFS.ReadDir("sql/migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := migrationFS.ReadFile("sql/migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute %s: %w", entry.Name(), err)
		}
		log.Printf("Migration applied: %s", entry.Name())
	}
	return nil
}
