package mymongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func init() {
	// connect to the database
	clientOptions := options.Client().ApplyURI("mongodb://mymongo-container:27017")
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	// check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(err)
	}
}

func GetMongoClient() *mongo.Client {
	return client
}
