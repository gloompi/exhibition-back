package schema

import (
	"errors"
	"fmt"
	"github.com/graphql-go/graphql"
	"online-exhibition.com/app/utils"
)

type User struct {
	UserId      string         `json:"user_id"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	UserName    string         `json:"user_name"`
	Email       string         `json:"email"`
	Phone       string	       `json:"phone"`
	DateOfBirth string         `json:"date_of_birth"`
	IsActive    bool           `json:"is_active"`
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

func readUsersSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(userType),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
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
			errCheck(err)

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
				errCheck(err)
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
				"status": &graphql.Field{Type: graphql.String},
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
			errCheck(err)
			defer stmt.Close()

			_, err = stmt.Exec()
			res := struct {
				Status string `json:"status"`
			}{
				Status: "bad",
			}

			if err == nil {
				res.Status = "ok"
			}

			return res, err
		},
	}
}

func readLoginUserSchema() *graphql.Field {
	return &graphql.Field{
		Type: userType,
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
			errCheck(err)

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
				errCheck(err)
			}

			correctPassword := utils.CheckPassword(existingPassword, password)
			if correctPassword == false {
				return nil, errors.New("wrong username or password")
			}

			return user, err
		},
	}
}
