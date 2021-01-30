package main

import (
	"fmt"
	"github.com/jstolwijk/simple-migrator"
	"path/filepath"
)

func main() {
	applyDatabaseMigrations()

	// Application code e.g. expose rest api
}

func applyDatabaseMigrations() {
	// use ./database/docker-compose.yml to spin up a "test" database
	url := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		"admin",
		"supersecretpassword123",
		"localhost",
		"5432",
		"postgres")

	path, _ := filepath.Abs("example/migrations")

	migrator.Migrate(url, path)
}
