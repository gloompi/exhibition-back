package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"io"
	"log"
	"net/http"
	"online-exhibition.com/app/dbConnection"
	schemaPkg "online-exhibition.com/app/schema"
	"online-exhibition.com/app/utils"
)

var conf config
var db *sql.DB

func init() {
	conf = readConfig()
	db = dbConnection.ReadConnection().DB
}

func main() {
	listenAt := fmt.Sprintf(":%d", conf.port)
	fmt.Println("DB-CONNECTION------", db.Ping())

	defer db.Close()

	// graphql
	schema, err := graphql.NewSchema(*schemaPkg.ReadSchema())
	if err != nil {
		log.Fatalln(err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	// route handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/generate-live-token", handleLiveToken)
	http.Handle("/graphql", corsMiddleware(requestMiddleware(h)))
	log.Printf("Open the following URL in the browser: http://localhost:%d\n", conf.port)
	log.Fatal(http.ListenAndServe(listenAt, nil))
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "You are doing Great!")
}

// Provide request instance through context
func requestMiddleware(next *handler.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := context.WithValue(req.Context(), "request", req)

		next.ContextHandler(ctx, w, req)
	})
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Origin,X-Requested-With,Content-Type,Accept")
		next.ServeHTTP(w, req)
	})
}

// Generate live token
func handleLiveToken(w http.ResponseWriter, req *http.Request) {
	userId, ok := req.URL.Query()["userId"]

	if !ok || len(userId[0]) < 1 {
		io.WriteString(w, "Please provide correct user id")
		return
	}

	td, err := utils.CreateLiveToken(userId[0])
	if err != nil {
		io.WriteString(w, "Failed while creating a token")
		return
	}

	err = utils.CreateAuth(userId[0], td)
	if err != nil {
		io.WriteString(w, "Failed while creating an auth")
		return
	}

	io.WriteString(w, "Everything is fine, here is your token "+td.AccessToken)
}
