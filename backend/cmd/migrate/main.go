package main

import (
	"errors"
	"flag"
	"log"
	"path/filepath"

	"expense-app/backend/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	path := flag.String("path", "migrations", "migration directory")
	flag.Parse()
	if flag.NArg() != 1 || (flag.Arg(0) != "up" && flag.Arg(0) != "down") {
		log.Fatal("usage: go run ./cmd/migrate [-path migrations] up|down (down rolls back one migration)")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	absPath, err := filepath.Abs(*path)
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.New("file://"+filepath.ToSlash(absPath), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("initialize migration: %v", err)
	}
	if flag.Arg(0) == "up" {
		err = m.Up()
	} else {
		err = m.Steps(-1)
	}
	sourceErr, databaseErr := m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration: %v", err)
	}
	if sourceErr != nil {
		log.Fatalf("close migration source: %v", sourceErr)
	}
	if databaseErr != nil {
		log.Fatalf("close migration database: %v", databaseErr)
	}
	log.Printf("migration %s complete", flag.Arg(0))
}
