package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

type connection struct {
	DB *sql.DB
}

var connectionInstance *connection

func ReadConnection() *connection {
	psqlUser := os.Getenv("PSQL_USERNAME")
	psqlPassword := os.Getenv("PSQL_PASSWORD")
	psqlAddress := os.Getenv("PSQL_ADDRESS")
	psqlDb := os.Getenv("PSQL_DB")

	if psqlAddress == "" {
		psqlAddress = "localhost:5432"
	}

	if psqlDb == "" {
		psqlDb = "streaming_service"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", psqlUser, psqlPassword, psqlAddress, psqlDb)

	if connectionInstance == nil {
		db, err := sql.Open("postgres", connStr)

		if err != nil {
			log.Fatalln(err)
		}
		connectionInstance = &connection{db}
	}

	return connectionInstance
}
