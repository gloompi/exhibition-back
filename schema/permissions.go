package schema

import (
	"fmt"
	"github.com/graphql-go/graphql"
)

// QUERIES
func readAdminsSchema() *graphql.Field {
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
				from admins
					inner join users using(user_id);
			`)
			rows, err := connection.DB.Query(query)
			errCheck(err)

			var admins []*User

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
				admins = append(admins, &user)
			}

			return admins, nil
		},
	}
}

func readProducersSchema() *graphql.Field {
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
				from producers
					inner join users using(user_id);
			`)
			rows, err := connection.DB.Query(query)
			errCheck(err)

			var producers []*User

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
				producers = append(producers, &user)
			}

			return producers, nil
		},
	}
}

func readAudienceSchema() *graphql.Field {
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
				from users u
				where 
					u.user_id not in
						(select user_id from admins)
					and u.user_id not in
						(select user_id from producers);
			`)
			rows, err := connection.DB.Query(query)
			errCheck(err)

			var audience []*User

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
				audience = append(audience, &user)
			}

			return audience, nil
		},
	}
}

// MUTATIONS
func readAddToAdminSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "AddToAdminResponse",
			Fields: graphql.Fields{
				"status": &graphql.Field{Type: graphql.String},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"userId": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			userId, _ := params.Args["userId"].(string)

			query := fmt.Sprintf(`
					insert into admins (user_id)
					values ('%v');
				`, userId)

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

func readAddToProducerSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "AddToProducerResponse",
			Fields: graphql.Fields{
				"status": &graphql.Field{Type: graphql.String},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"userId": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			userId, _ := params.Args["userId"].(string)

			query := fmt.Sprintf(`
					insert into producers (user_id)
					values ('%v');
				`, userId)

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
