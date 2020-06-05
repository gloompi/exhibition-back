package schema

import (
	"github.com/gloompi/tantora-back/app/dbConnection"
	"github.com/graphql-go/graphql"
)

var connection = dbConnection.ReadConnection()

func ReadSchema() *graphql.SchemaConfig {
	schemaConfig := graphql.SchemaConfig{
		Query:    rootQuery(),
		Mutation: rootMutation(),
	}

	return &schemaConfig
}

func rootQuery() *graphql.Object {
	fields := graphql.Fields{
		"me":          readMeSchema(),
		"users":       readUsersSchema(),
		"producers":   readProducersSchema(),
		"audience":    readAudienceSchema(),
		"exhibitions": readExhibitionsSchema(),
		"loginUser":   readLoginUserSchema(),
		"admins":      readAdminsSchema(),
		"logout":      readLogoutSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})
}

func rootMutation() *graphql.Object {
	fields := graphql.Fields{
		"createUser":       readCreateUserSchema(),
		"refreshToken":     readRefreshTokenSchema(),
		"createExhibition": readCreateExhibitionSchema(),
		"addToAdmins":      readAddToAdminSchema(),
		"addToProducer":    readAddToProducerSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootMutation", Fields: fields})
}
