package main

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"io"
	"log"
	"net/http"
	"online-exhibition.com/app/dbConnection"
	schemaPkg "online-exhibition.com/app/schema"
)

var conf config = readConfig()

func main() {
	connection := dbConnection.ReadConnection()
	listenAt := fmt.Sprintf(":%d", conf.port)
	fmt.Println("CONNECTION------", connection.DB.Ping())
	defer connection.DB.Close()

	// graphql
	schema, err := graphql.NewSchema(*schemaPkg.ReadSchema())
	errCheck(err)
	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	// http
	http.HandleFunc("/", handleIndex)
	http.Handle("/graphql", h)
	log.Printf("Open the following URL in the browser: http://localhost:%d\n", conf.port)
	log.Fatal(http.ListenAndServe(listenAt, nil))
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Hello World!")
}

func errCheck(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
