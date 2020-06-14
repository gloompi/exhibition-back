package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gloompi/tantora-back/app/dbConnection"
	grpcServer "github.com/gloompi/tantora-back/app/grpc"
	"github.com/gloompi/tantora-back/app/proto/tantorapb"
	schemaPkg "github.com/gloompi/tantora-back/app/schema"
	"github.com/gloompi/tantora-back/app/utils"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

var conf config
var db *sql.DB

func init() {
	conf = readConfig()
	db = dbConnection.ReadConnection().DB
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lisCh := make(chan net.Listener, 1)
	grpcSCh := make(chan *grpc.Server, 1)

	go initGRPCServer(lisCh, grpcSCh)
	go initHttpServer()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	lis := <-lisCh
	s := <-grpcSCh

	fmt.Println("\nStopping the app...")
	s.Stop()
	lis.Close()
	fmt.Println("Everything is closed properly!")
}

// HTTP Server
func initHttpServer() {
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
	http.HandleFunc("/generate-live-token", handleLiveToken)
	http.Handle("/graphql", corsMiddleware(requestMiddleware(h)))
	log.Printf("Open the following URL in the browser: http://localhost:%d\n", conf.port)
	log.Fatal(http.ListenAndServe(listenAt, nil))
}

// GRPC Server
func initGRPCServer(lisCh chan<- net.Listener, grpcSCh chan<- *grpc.Server) {
	fmt.Println("Starting GRPC server")

	lis, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(conf.grpcPort))
	if err != nil {
		log.Fatalf("Failed to  listen: %v", err)
	}

	lisCh <- lis

	tls := false
	opts := []grpc.ServerOption{}

	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed loading certificates: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	tantorapb.RegisterChatServiceServer(s, &grpcServer.Server{})

	grpcSCh <- s

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
