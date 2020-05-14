package schema

import (
	"github.com/graphql-go/graphql"
)

func ReadPublicSchema() *graphql.SchemaConfig {
	schemaConfig := graphql.SchemaConfig{
		Query:    publicRootQuery(),
		Mutation: publicRootMutation(),
	}

	return &schemaConfig
}

func publicRootQuery() *graphql.Object {
	fields := graphql.Fields{
		"users":       readUsersSchema(),
		"producers":   readProducersSchema(),
		"audience":    readAudienceSchema(),
		"exhibitions": readExhibitionsSchema(),
		"loginUser":   readLoginUserSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})
}

func publicRootMutation() *graphql.Object {
	fields := graphql.Fields{
		"createUser":   readCreateUserSchema(),
		"refreshToken": readRefreshTokenSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootMutation", Fields: fields})
}
