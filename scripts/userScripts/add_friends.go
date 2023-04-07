package userScripts

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/alexander-winters/SENG468-A2/scripts/db"
	"go.mongodb.org/mongo-driver/bson"
)

// AddRandomFriends reads usernames from "users.txt" and adds a random list of friends for each user.
func AddRandomFriends() {
	// open "users.txt" file for reading
	file, err := os.Open("users.txt")
	if err != nil {
		log.Fatalf("Error opening users.txt: %v", err)
	}
	defer file.Close()

	// create a scanner to read lines from the file
	scanner := bufio.NewScanner(file)

	// create a slice to hold the usernames
	usernames := []string{}

	// read each line and extract the username, add it to the slice
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ", ")
		username := parts[2]
		usernames = append(usernames, username)
	}

	// check if there was an error during scanning
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error scanning users.txt: %v", err)
	}

	// get the MongoDB collection for the users
	usersCollection := db.GetMongoClient().Database("seng468-a2-db").Collection("users")

	// create a new context with a 10 second timeout and defer cancelling it
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// create a new source of random numbers
	src := rand.NewSource(time.Now().UnixNano())
	rand.New(src)

	// loop through each username in the slice
	for _, username := range usernames {
		// generate a random number of friends between 1 and 10
		numOfFriends := rand.Intn(len(usernames)-1) + 1
		friendList := []string{}

		// add friends to the list until the desired number is reached
		for len(friendList) < numOfFriends {
			friendIndex := rand.Intn(len(usernames))
			friend := usernames[friendIndex]

			// check if the friend is not the user and is not already in the list
			if friend != username && !contains(friendList, friend) {
				friendList = append(friendList, friend)
			}
		}

		// update the user's list of friends in MongoDB
		filter := bson.M{"username": username}
		update := bson.M{"$set": bson.M{"list_of_friends": friendList}}
		_, err := usersCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Error updating user's friends: %v", err)
		} else {
			fmt.Printf("Updated friends for %s: %v\n", username, friendList)
		}
	}
}

// contains checks if a string slice contains a given string item
func contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
