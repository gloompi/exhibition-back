package schema

import (
	"github.com/graphql-go/graphql"
	"log"
	"online-exhibition.com/app/dbConnection"
)

var connection = dbConnection.ReadConnection()

func ReadPrivateSchema() *graphql.SchemaConfig {
	schemaConfig := graphql.SchemaConfig{
		Query:    privateRootQuery(),
		Mutation: privateRootMutation(),
	}

	return &schemaConfig
}

func privateRootQuery() *graphql.Object {
	fields := graphql.Fields{
		"admins": readAdminsSchema(),
		"logout": readLogoutSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})
}

func privateRootMutation() *graphql.Object {
	fields := graphql.Fields{
		"createExhibition": readCreateExhibitionSchema(),
		"addToAdmins":      readAddToAdminSchema(),
		"addToProducer":    readAddToProducerSchema(),
	}

	return graphql.NewObject(graphql.ObjectConfig{Name: "RootMutation", Fields: fields})
}

func errCheck(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
