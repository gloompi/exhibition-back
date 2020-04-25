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
	fmt.Println("DB-CONNECTION------", connection.DB.Ping())

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
	http.Handle("/graphql", CorsMiddleware(h))
	log.Printf("Open the following URL in the browser: http://localhost:%d\n", conf.port)
	log.Fatal(http.ListenAndServe(listenAt, nil))
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "You are doing Great!")
}

// CORS middleware
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		next.ServeHTTP(w,r)
	})
}

func errCheck(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
