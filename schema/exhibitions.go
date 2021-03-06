package schema

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gloompi/tantora-back/app/utils"
	"github.com/graphql-go/graphql"
	"net/http"
	"strings"
)

type Exhibition struct {
	ExhibitionId string `json:"exhibition_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	StartDate    string `json:"start_date,omitempty"`
	CreatedDate  string `json:"created_date,omitempty"`
	OwnerId      string `json:"owner_id,omitempty"`
}

var exhibitionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Exhibition",
	Fields: graphql.Fields{
		"exhibitionId": &graphql.Field{Type: graphql.String},
		"name":         &graphql.Field{Type: graphql.String},
		"description":  &graphql.Field{Type: graphql.String},
		"startDate":    &graphql.Field{Type: graphql.String},
		"createdDate":  &graphql.Field{Type: graphql.String},
		"owner": &graphql.Field{
			Type: userType,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				exhibition, ok := params.Source.(*Exhibition)

				if !ok {
					return nil, errors.New("were not able to get the exhibition")
				}

				query := fmt.Sprintf(`
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
					where u.user_id = %v;
				`, exhibition.OwnerId)

				rows, err := connection.DB.Query(query)
				if err != nil {
					return nil, err
				}

				var user User

				for rows.Next() {
					err := rows.Scan(
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
				}

				return user, nil
			},
		},
	},
})

func readExhibitionSchema() *graphql.Field {
	return &graphql.Field{
		Type: exhibitionType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{Type: graphql.String},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			id, ok := params.Args["id"].(string)

			if !ok {
				return nil, errors.New("ID is required")
			}

			query := fmt.Sprintf(`
				select
					exhibition_id,
					name,
					description,
					start_date,
					created_date,
					owner_id
				from exhibitions ex
				where ex.exhibition_id = %v;
			`, id)

			rows, err := connection.DB.Query(query)
			if err != nil {
				return nil, err
			}

			exhibition := &Exhibition{}

			for rows.Next() {
				err = rows.Scan(
					&exhibition.ExhibitionId,
					&exhibition.Name,
					&exhibition.Description,
					&exhibition.StartDate,
					&exhibition.CreatedDate,
					&exhibition.OwnerId,
				)
				if err != nil {
					return nil, err
				}
			}

			decodedStr, _ := hex.DecodeString(exhibition.Description)
			exhibition.Description = string(decodedStr)
			return exhibition, nil
		},
	}
}

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
					owner_id
				from exhibitions ex
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
					&exhibition.OwnerId,
				)
				if err != nil {
					return nil, err
				}

				decodedStr, _ := hex.DecodeString(exhibition.Description)
				exhibition.Description = string(decodedStr)
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
			`, name, hex.EncodeToString([]byte(description)), startDate, ownerId)

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
