package main

import (
	"flag"
	"fmt"

	"github.com/alexander-winters/SENG468-A2/scripts/userScripts"
)

func main() {
	// Define the command-line flag
	createUsers := flag.Int("c", 0, "Number of users to create")

	// Parse the command-line flags
	flag.Parse()

	// Create users if the flag is set
	if *createUsers > 0 {
		fmt.Println("Creating users...")
		userScripts.CreateRandomUsers(*createUsers)
	}

	// Download users
	fmt.Println("Downloading users...")
	userScripts.DownloadUsers()

	// Add random friends
	fmt.Println("Adding random friends...")
	userScripts.AddRandomFriends()
}
