package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"online-exhibition.com/app/dbConnection"
	schemaPkg "online-exhibition.com/app/schema"
	"online-exhibition.com/app/utils"
	"os"
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

	playground := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: false,
	})

	// route handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/token/refresh", handleRefresh)
	http.Handle("/playground", playground)
	http.Handle("/graphql", CorsMiddleware(TokenRequiredMiddleware(h)))
	log.Printf("Open the following URL in the browser: http://localhost:%d\n", conf.port)
	log.Fatal(http.ListenAndServe(listenAt, nil))
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "You are doing Great!")
}

func handleLogin(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var userCredentials struct{
		UserName string
		Password string
	}
	body, _ := ioutil.ReadAll(req.Body)
	err := json.Unmarshal(body, &userCredentials)

	if err != nil {
		http.Error(w, "parsing json failed", http.StatusInternalServerError)
		return
	}

	query := fmt.Sprintf(`
		select
			user_id,
			user_name,
			password,
			first_name,
			last_name,
			email,
			phone,
			date_of_birth,
			is_active
		from users
		where user_name = '%v';
	`, userCredentials.UserName)
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Login DB query failed", http.StatusInternalServerError)
	}

	var existingPassword []byte
	var user schemaPkg.User

	for rows.Next() {
		err = rows.Scan(
			&user.UserId,
			&user.UserName,
			&existingPassword,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Phone,
			&user.DateOfBirth,
			&user.IsActive,
		)
		errCheck(w, err, http.StatusInternalServerError)
	}

	correctPassword := utils.CheckPassword(existingPassword, userCredentials.Password)
	if correctPassword == false {
		http.Error(w, "Password is not correct", http.StatusUnauthorized)
		return
	}

	ts, err := utils.CreateToken(user.UserId)
	if err != nil {
		http.Error(w, "Token generation is failed", http.StatusUnauthorized)
		return
	}

	saveErr := utils.CreateAuth(user.UserId, ts)
	if saveErr != nil {
		http.Error(w, "Token Expired", http.StatusUnauthorized)
		return
	}

	res := schemaPkg.Token{
		ts.AccessToken,
		ts.RefreshToken,
	}

	b, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "Marshal failed", http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))
}

func handleLogout(w http.ResponseWriter, req *http.Request) {
	au, err := utils.ExtractTokenMetadata(req)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	deleted, delErr := utils.DeleteAuth(au.AccessUuid)
	if delErr != nil || deleted == 0 { //if any goes wrong
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	io.WriteString(w, "Successfully logged out")
}


func handleRefresh(w http.ResponseWriter, req *http.Request) {
	mapToken := map[string]string{}
	refreshToken := mapToken["refresh_token"]

	//verify the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		userId, err := claims["user_id"].(string)
		if err != false {
			http.Error(w, "Error occurred", http.StatusUnprocessableEntity)
			return
		}
		//Delete the previous Refresh Token
		deleted, delErr := utils.DeleteAuth(refreshUuid)
		if delErr != nil || deleted == 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := utils.CreateToken(userId)
		if  createErr != nil {
			http.Error(w, createErr.Error(), http.StatusForbidden)
			return
		}
		//save the tokens metadata to redis
		saveErr := utils.CreateAuth(userId, ts)
		if saveErr != nil {
			http.Error(w, saveErr.Error(), http.StatusForbidden)
			return
		}

		res := schemaPkg.Token{
			ts.AccessToken,
			ts.RefreshToken,
		}

		b, marshalErr := json.Marshal(res)
		if marshalErr != nil {
			http.Error(w, "Marshal failed", http.StatusInternalServerError)
			return
		}

		io.WriteString(w, string(b))
	} else {
		http.Error(w, "refresh expired", http.StatusUnauthorized)
	}
}

func TokenRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tokenAuth, err := utils.ExtractTokenMetadata(req)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		_, err = utils.FetchAuth(tokenAuth)
		if err != nil {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := utils.TokenValid(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
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
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		next.ServeHTTP(w, req)
	})
}

func errCheck(w http.ResponseWriter, err error, status int) {
	if err != nil {
		http.Error(w, err.Error(), status)
	}
}
