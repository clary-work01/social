package main

import (
	"log"

	"github.com/chainflow/chainflow-api/internal/db"
	"github.com/chainflow/chainflow-api/internal/env"
	"github.com/chainflow/chainflow-api/internal/store"
)

func main() {
	conn, err := db.New(
		env.GetEnvString("DB_DSN", "postgres://admin:admin123@localhost/chainflow_db?sslmode=disable"),
		3,
		3,
		"15m",
	)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewStorage(conn)

	db.Seed(store, conn)
}
