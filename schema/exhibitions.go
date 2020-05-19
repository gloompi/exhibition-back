package schema

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"net/http"
	"online-exhibition.com/app/utils"
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
		Args: graphql.FieldConfigArgument{
			"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
			"offset": &graphql.ArgumentConfig{Type: graphql.Int},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			limit, ok := params.Args["limit"].(int)
			offset, _ := params.Args["offset"].(int)

			queryString := `
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
						on ex.owner_id = u.user_id
				order by created_date desc
			`

			query := fmt.Sprintf(queryString+`
				limit %v offset %v;
			`, limit, offset)

			if !ok {
				query = queryString
			}

			rows, err := connection.DB.Query(query)
			if err != nil {
				return nil, err
			}

			var exhibitions []*Exhibition

			for rows.Next() {
				var exhibition Exhibition

				err = rows.Scan(
					&exhibition.ExhibitionId,
					&exhibition.Name,
					&exhibition.Description,
					&exhibition.StartDate,
					&exhibition.CreatedDate,
					&exhibition.Owner.FirstName,
					&exhibition.Owner.LastName,
					&exhibition.Owner.Email,
					&exhibition.Owner.DateOfBirth,
					&exhibition.Owner.IsActive,
					&exhibition.Owner.UserId,
					&exhibition.Owner.Phone,
					&exhibition.Owner.UserName,
				)
				if err != nil {
					return nil, err
				}

				exhibitions = append(exhibitions, &exhibition)
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
			req := params.Context.Value("request").(*http.Request)
			_, err := utils.TokenValid(req)
			if err != nil {
				return nil, err
			}

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
			if err != nil {
				return nil, err
			}

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
