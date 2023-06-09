package userScripts

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexander-winters/SENG468-A2/mymongo/models"
	"github.com/alexander-winters/SENG468-A2/scripts/db"
	"go.mongodb.org/mongo-driver/bson"
)

func DownloadUsersToFile(filename string) {
	// Initialize the MongoDB client
	db.GetMongoClient()

	// Get a handle to the users collection
	usersCollection := db.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all users in the collection
	cursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Error finding users: %v", err)
	}
	defer cursor.Close(ctx)

	// Open a file with the specified filename for writing
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	// Iterate through the users and save the required information to the file
	var user models.User
	userNumber := 1
	for cursor.Next(ctx) {
		err := cursor.Decode(&user)
		if err != nil {
			log.Printf("Error decoding user: %v", err)
			continue
		}

		line := fmt.Sprintf("%d, %s, %s\n", userNumber, user.ID.Hex(), user.Username)
		if _, err := file.WriteString(line); err != nil {
			log.Printf("Error writing to file: %v", err)
		} else {
			userNumber++
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Error with cursor: %v", err)
	}

	fmt.Printf("Users saved to %s\n", filename)
}
