package schema

import (
	"github.com/graphql-go/graphql"
	"log"
	"online-exhibition.com/app/dbConnection"
)

var connection = dbConnection.ReadConnection()

func ReadSchema() *graphql.SchemaConfig {
	schemaConfig := graphql.SchemaConfig{
		Query:    readRootQuery(),
		Mutation: readRootMutation(),
	}

	return &schemaConfig
}

func readRootQuery() *graphql.Object {
	fields := graphql.Fields{
		"users":     readUsersSchema(),
		"loginUser": readLoginUserSchema(),
		"admins":    readAdminsSchema(),
		"producers": readProducersSchema(),
		"audience":  readAudienceSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})
}

func readRootMutation() *graphql.Object {
	fields := graphql.Fields{
		"createUser":    readCreateUserSchema(),
		"addToAdmins":   readAddToAdminSchema(),
		"addToProducer": readAddToProducerSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootMutation", Fields: fields})
}

func errCheck(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
