package commentScripts

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

func DownloadCommentsToFile(filename string) {
	// Initialize the MongoDB client
	db.GetMongoClient()

	// Get a handle to the comments collection
	commentsCollection := db.GetMongoClient().Database("seng468-a2-db").Collection("comments")

	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all comments in the collection
	cursor, err := commentsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Error finding comments: %v", err)
	}
	defer cursor.Close(ctx)

	// Open a file with the specified filename for writing
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	// Iterate through the comments and save the required information to the file
	var comment models.Comment
	commentNumber := 1
	for cursor.Next(ctx) {
		err := cursor.Decode(&comment)
		if err != nil {
			log.Printf("Error decoding comment: %v", err)
			continue
		}

		line := fmt.Sprintf("%d, %s, %d, %s\n", commentNumber, comment.Username, comment.PostNumber, comment.ID.Hex())
		if _, err := file.WriteString(line); err != nil {
			log.Printf("Error writing to file: %v", err)
		} else {
			commentNumber++
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Error with cursor: %v", err)
	}

	fmt.Printf("Comments saved to %s\n", filename)
}
