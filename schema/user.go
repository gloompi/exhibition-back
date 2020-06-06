package schema

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gloompi/tantora-back/app/utils"
	"github.com/graphql-go/graphql"
	"net/http"
	"os"
)

type User struct {
	UserId      string `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	UserName    string `json:"user_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth"`
	IsActive    bool   `json:"is_active"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"userId":      &graphql.Field{Type: graphql.String},
		"firstName":   &graphql.Field{Type: graphql.String},
		"lastName":    &graphql.Field{Type: graphql.String},
		"userName":    &graphql.Field{Type: graphql.String},
		"email":       &graphql.Field{Type: graphql.String},
		"phone":       &graphql.Field{Type: graphql.String},
		"dateOfBirth": &graphql.Field{Type: graphql.String},
		"isActive":    &graphql.Field{Type: graphql.Boolean},
	},
})

var tokenType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Token",
	Fields: graphql.Fields{
		"accessToken":  &graphql.Field{Type: graphql.String},
		"refreshToken": &graphql.Field{Type: graphql.String},
	},
})

func readMeSchema() *graphql.Field {
	return &graphql.Field{
		Type: userType,
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			req := params.Context.Value("request").(*http.Request)
			userId, err := utils.TokenValid(req)

			if err != nil {
				return nil, err
			}

			query := fmt.Sprintf(`
				select
					user_id,
					user_name,
					first_name,
					last_name,
					email,
					phone,
					date_of_birth,
					is_active
				from users u where u.user_id = '%v';
			`, userId)

			rows, err := connection.DB.Query(query)
			if err != nil {
				return nil, err
			}

			var user User

			for rows.Next() {
				err = rows.Scan(
					&user.UserId,
					&user.UserName,
					&user.FirstName,
					&user.LastName,
					&user.Email,
					&user.Phone,
					&user.DateOfBirth,
					&user.IsActive,
				)

				if err != nil {
					return nil, err
				}
			}

			return user, nil
		},
	}
}

func readUsersSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(userType),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			req := params.Context.Value("request").(*http.Request)
			_, err := utils.TokenValid(req)
			if err != nil {
				return nil, err
			}

			query := fmt.Sprintln(`
				select
					first_name,
					last_name,
					email,
					date_of_birth,
					is_active,
					user_id,
					phone,
					user_name
				from users;
			`)
			rows, err := connection.DB.Query(query)
			if err != nil {
				return nil, err
			}

			var users []*User

			for rows.Next() {
				var user User
				err = rows.Scan(
					&user.FirstName,
					&user.LastName,
					&user.Email,
					&user.DateOfBirth,
					&user.IsActive,
					&user.UserId,
					&user.Phone,
					&user.UserName,
				)
				if err != nil {
					return nil, err
				}

				users = append(users, &user)
			}

			return users, nil
		},
	}
}

func readCreateUserSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "CreateUserResponse",
			Fields: graphql.Fields{
				"user":  &graphql.Field{Type: userType},
				"token": &graphql.Field{Type: tokenType},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"userId":      &graphql.ArgumentConfig{Type: graphql.String},
			"firstName":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			"lastName":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			"userName":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			"email":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			"password":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			"phone":       &graphql.ArgumentConfig{Type: graphql.String},
			"dateOfBirth": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			"isActive":    &graphql.ArgumentConfig{Type: graphql.Boolean},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			firstName, _ := params.Args["firstName"].(string)
			lastName, _ := params.Args["lastName"].(string)
			userName, _ := params.Args["userName"].(string)
			email, _ := params.Args["email"].(string)
			password, _ := params.Args["password"].(string)
			phone, _ := params.Args["phone"].(string)
			dateOfBirth, _ := params.Args["dateOfBirth"].(string)
			isActive, _ := params.Args["isActive"].(bool)

			hashedPassword, _ := utils.EncryptPassword(password)

			query := fmt.Sprintf(`
					insert into users (first_name, last_name, email, date_of_birth, is_active, phone, "password", user_name)
					values
						('%v', '%v', '%v', '%v', %v, '%v', '%v', '%v');
				`, firstName, lastName, email, dateOfBirth, isActive, phone, string(hashedPassword), userName)

			stmt, err := connection.DB.Prepare(query)
			if err != nil {
				return nil, err
			}

			defer stmt.Close()

			_, err = stmt.Exec()

			if err != nil {
				return nil, err
			}

			query = fmt.Sprintf(`
				select
					user_id,
					user_name,
					first_name,
					last_name,
					email,
					phone,
					date_of_birth,
					is_active
				from users u where u.user_name = '%v';
			`, userName)

			rows, err := connection.DB.Query(query)
			if err != nil {
				return nil, err
			}

			var user User

			for rows.Next() {
				err = rows.Scan(
					&user.UserId,
					&user.UserName,
					&user.FirstName,
					&user.LastName,
					&user.Email,
					&user.Phone,
					&user.DateOfBirth,
					&user.IsActive,
				)

				if err != nil {
					return nil, err
				}
			}

			ts, err := utils.CreateToken(user.UserId)
			if err != nil {
				return nil, err
			}

			err = utils.CreateAuth(user.UserId, ts)
			if err != nil {
				return nil, err
			}

			res := struct {
				User  User
				Token Token
			}{
				user,
				Token{
					ts.AccessToken,
					ts.RefreshToken,
				},
			}

			return res, nil
		},
	}
}

func readLoginUserSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "LoginResponse",
			Fields: graphql.Fields{
				"user":  &graphql.Field{Type: userType},
				"token": &graphql.Field{Type: tokenType},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"userName": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			"password": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			userName, _ := params.Args["userName"].(string)
			password, _ := params.Args["password"].(string)

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
			`, userName)
			rows, err := connection.DB.Query(query)
			if err != nil {
				return nil, err
			}

			var existingPassword []byte
			var user User

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
				if err != nil {
					return nil, err
				}
			}

			correctPassword := utils.CheckPassword(existingPassword, password)
			if correctPassword == false {
				return nil, errors.New("wrong username or password")
			}
			ts, err := utils.CreateToken(user.UserId)
			if err != nil {
				return nil, err
			}

			err = utils.CreateAuth(user.UserId, ts)
			if err != nil {
				return nil, err
			}

			res := struct {
				User  User
				Token Token
			}{
				user,
				Token{
					ts.AccessToken,
					ts.RefreshToken,
				},
			}

			return res, nil
		},
	}
}

func readLogoutSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "LogoutResponse",
			Fields: graphql.Fields{
				"deleted": &graphql.Field{Type: graphql.Int},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"token": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			req := params.Context.Value("request").(*http.Request)
			_, err := utils.TokenValid(req)
			if err != nil {
				return nil, err
			}

			token, _ := params.Args["token"].(string)

			au, err := utils.ExtractTokenMetadataString(token)
			if err != nil {
				return nil, err
			}

			deleted, err := utils.DeleteAuth(au.AccessUuid)
			if err != nil || deleted == 0 {
				return nil, err
			}

			return struct {
				Deleted int64
			}{
				deleted,
			}, nil
		},
	}
}

func readRefreshTokenSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "RefreshTokenResponse",
			Fields: graphql.Fields{
				"token": &graphql.Field{Type: tokenType},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"refreshToken": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			refreshToken, _ := params.Args["refreshToken"].(string)

			token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(os.Getenv("REFRESH_SECRET")), nil
			})

			if err != nil {
				return nil, err
			}

			if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
				return nil, errors.New("unauthorized")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if ok && token.Valid {
				refreshUuid, ok := claims["refresh_uuid"].(string)
				if !ok {
					return nil, errors.New("failed to get `refresh_uuid`")
				}

				userId, ok := claims["user_id"].(string)
				if !ok {
					return nil, errors.New("failed to get `user_id`")
				}

				deleted, err := utils.DeleteAuth(refreshUuid)
				if err != nil || deleted == 0 {
					return nil, errors.New("failed to delete old `refresh token`, it probably were deleted already")
				}

				ts, err := utils.CreateToken(userId)
				if err != nil {
					return nil, errors.New("failed to create new token")
				}

				err = utils.CreateAuth(userId, ts)
				if err != nil {
					return nil, err
				}

				res := struct {
					Token Token
				}{
					Token{
						ts.AccessToken,
						ts.RefreshToken,
					},
				}

				return res, nil
			} else {
				return nil, errors.New("refresh token expired")
			}
		},
	}
}
