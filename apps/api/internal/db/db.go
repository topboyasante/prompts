package db

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("database URL is required")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open gorm connection: %w", err)
	}

	return db, nil
}

func RunMigrations(databaseURL, migrationsPath string) error {
	log.Printf("db: checking migrations in %s", migrationsPath)

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}
	if len(entries) == 0 {
		log.Printf("db: no migration files found, skipping")
		return nil
	}

	log.Printf("db: running %d migration file(s)", len(entries))

	migrator, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("db: migrations already up to date")
			srcErr, dbErr := migrator.Close()
			if srcErr != nil {
				return srcErr
			}
			if dbErr != nil {
				return dbErr
			}
			return nil
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	srcErr, dbErr := migrator.Close()
	if srcErr != nil {
		return srcErr
	}
	if dbErr != nil {
		return dbErr
	}

	log.Printf("db: migrations applied successfully")

	return nil
}
