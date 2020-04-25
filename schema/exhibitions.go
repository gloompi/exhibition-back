package schema

import (
	"database/sql"
	"fmt"
	"github.com/graphql-go/graphql"
	"strings"
)

type Exhibition struct {
	ExhibitionId string `json:"exhibition_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	StartDate    string `json:"start_date"`
	CreatedDate  string `json:"created_date"`
	Owner        User   `json:"owner"`
}

var exhibitionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Exhibition",
	Fields: graphql.Fields{
		"exhibitionId": &graphql.Field{Type: graphql.String},
		"name":         &graphql.Field{Type: graphql.String},
		"description":  &graphql.Field{Type: graphql.String},
		"startDate":    &graphql.Field{Type: graphql.String},
		"createdDate":  &graphql.Field{Type: graphql.String},
		"owner":        &graphql.Field{Type: userType},
	},
})

func readExhibitionsSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewList(exhibitionType),
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			query := fmt.Sprintln(`
				select
					exhibition_id,
					name,
					description,
					start_date,
					created_date,
					first_name,
					last_name,
					email,
					date_of_birth,
					is_active,
					user_id,
					phone,
					user_name
				from exhibitions ex
					inner join users u
						on ex.owner_id = u.user_id;
			`)
			rows, err := connection.DB.Query(query)
			errCheck(err)

			var exhibitions []*Exhibition
			var exhibitionId, name, description, startDate, createdDate, firstName, lastName, email, dateOfBirth, userId, userName string
			var phone sql.NullString
			var isActive bool

			for rows.Next() {
				err = rows.Scan(
					&exhibitionId,
					&name,
					&description,
					&startDate,
					&createdDate,
					&firstName,
					&lastName,
					&email,
					&dateOfBirth,
					&isActive,
					&userId,
					&phone,
					&userName,
				)
				errCheck(err)

				exhibitions = append(exhibitions, &Exhibition{
					exhibitionId,
					name,
					string(description),
					startDate,
					createdDate,
					User{
						userId,
						firstName,
						lastName,
						userName,
						email,
						phone,
						dateOfBirth,
						isActive,
					},
				})
			}

			return exhibitions, nil
		},
	}
}

func readCreateExhibitionSchema() *graphql.Field {
	return &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "CreateExhibitionResponse",
			Fields: graphql.Fields{
				"status": &graphql.Field{Type: graphql.String},
			},
		}),
		Args: graphql.FieldConfigArgument{
			"name":        &graphql.ArgumentConfig{Type: graphql.String},
			"description": &graphql.ArgumentConfig{Type: graphql.String},
			"startDate":   &graphql.ArgumentConfig{Type: graphql.String},
			"ownerId":     &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			name, _ := params.Args["name"].(string)
			description, _ := params.Args["description"].(string)
			startDate, _ := params.Args["startDate"].(string)
			ownerId, _ := params.Args["ownerId"].(string)
			description = strings.Replace(description, "'", "''", -1)

			query := fmt.Sprintf(`
					insert into exhibitions (name, description, start_date, owner_id)
					values ('%v', '%v', '%v', '%v');
				`, name, description, startDate, ownerId)

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
