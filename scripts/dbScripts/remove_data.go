package dbScripts

import (
	"context"
	"fmt"
	"log"

	"github.com/alexander-winters/SENG468-A2/scripts/db"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
)

var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func RemoveDBData() {
	database := db.GetMongoClient().Database("seng468-a2-db")

	// Get the list of collection names
	collectionNames, err := database.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through each collection and delete all documents
	for _, collectionName := range collectionNames {
		collection := database.Collection(collectionName)
		res, err := collection.DeleteMany(context.Background(), bson.M{})

		if err != nil {
			log.Fatalf("Error deleting documents in collection '%s': %v", collectionName, err)
		}

		fmt.Printf("Deleted %d documents from collection '%s'\n", res.DeletedCount, collectionName)
	}

	fmt.Println("All documents deleted from all collections.")
}

func RemoveRedisData() {
	// Check the connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	// Delete all keys (documents) from Redis
	err = rdb.FlushDB(context.Background()).Err()
	if err != nil {
		log.Fatalf("Error deleting data from Redis: %v", err)
	}

	fmt.Println("All data deleted from Redis.")
}
