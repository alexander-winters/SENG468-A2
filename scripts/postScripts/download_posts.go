package postScripts

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexander-winters/SENG468-A2/mymongo"
	"github.com/alexander-winters/SENG468-A2/mymongo/models"
	"go.mongodb.org/mongo-driver/bson"
)

func DownloadPostsToFile(filename string) {
	// Initialize the MongoDB client
	mymongo.GetMongoClient()

	// Get a handle to the posts collection
	postsCollection := mymongo.GetMongoClient().Database("seng468-a2-db").Collection("posts")

	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all posts in the collection
	cursor, err := postsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Error finding posts: %v", err)
	}
	defer cursor.Close(ctx)

	// Open a file with the specified filename for writing
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	// Iterate through the posts and save the required information to the file
	var post models.Post
	postNumber := 1
	for cursor.Next(ctx) {
		err := cursor.Decode(&post)
		if err != nil {
			log.Printf("Error decoding post: %v", err)
			continue
		}

		line := fmt.Sprintf("%d, %s, %d\n", postNumber, post.Username, post.PostNumber)
		if _, err := file.WriteString(line); err != nil {
			log.Printf("Error writing to file: %v", err)
		} else {
			postNumber++
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Error with cursor: %v", err)
	}

	fmt.Printf("Posts saved to %s\n", filename)
}
