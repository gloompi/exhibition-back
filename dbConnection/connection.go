package dbConnection

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
	psql_user := os.Getenv("PSQL_USERNAME")
	psql_password := os.Getenv("PSQL_PASSWORD")
	psql_address := os.Getenv("PSQL_ADDRESS")

	if psql_address == "" {
		psql_address = "localhost:5432"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s/streaming_service?sslmode=disable", psql_user, psql_password, psql_address)

	if connectionInstance == nil {
		db, err := sql.Open("postgres", connStr)

		if err != nil {
			log.Fatalln(err)
		}
		connectionInstance = &connection{db}
	}

	return connectionInstance
}
