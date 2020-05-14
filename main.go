package main

import (
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
	privateSchema, err := graphql.NewSchema(*schemaPkg.ReadPrivateSchema())
	if err != nil {
		log.Fatalln(err)
	}

	publicSchema, err := graphql.NewSchema(*schemaPkg.ReadPublicSchema())
	if err != nil {
		log.Fatalln(err)
	}

	privatePlayground := handler.New(&handler.Config{
		Schema:     &privateSchema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	publicHandler := handler.New(&handler.Config{
		Schema:     &publicSchema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	privateHandler := handler.New(&handler.Config{
		Schema:     &privateSchema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: false,
	})

	// route handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/generate-live-token", handleLiveToken)
	http.Handle("/private-playground", privatePlayground)
	http.Handle("/graphql", CorsMiddleware(publicHandler))
	http.Handle("/graphql/private", CorsMiddleware(TokenAuthMiddleware(privateHandler)))
	log.Printf("Open the following URL in the browser: http://localhost:%d\n", conf.port)
	log.Fatal(http.ListenAndServe(listenAt, nil))
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "You are doing Great!")
}

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

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// avoid auth check for OPTION method
		if req.Method == http.MethodOptions {
			next.ServeHTTP(w, req)
			return
		}

		err := utils.TokenValid(req)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// CORS middleware
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Origin,X-Requested-With,Content-Type,Accept")
		next.ServeHTTP(w, req)
	})
}
