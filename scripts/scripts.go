package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alexander-winters/SENG468-A2/scripts/commentScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/dbScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/postScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/userScripts"
)

func main() {
	// Define the command-line flags
	createUsers := flag.Int("c", 0, "Number of users to create")
	deleteData := flag.Bool("d", false, "Delete all data (requires confirmation)")
	confirmDelete := flag.Bool("y", false, "Skip confirmation prompt when using -d")
	addFriends := flag.Bool("af", false, "Add random friends to users")
	help := flag.Bool("h", false, "Display help information")
	helpLong := flag.Bool("help", false, "Display help information")

	// Parse the command-line flags
	flag.Parse()

	if *help || *helpLong {
		displayHelp()
		return
	}

	// Download users
	filename := "users.txt"
	fmt.Println("Downloading users...")
	userScripts.DownloadUsersToFile(filename)

	// Create users if the flag is set
	if *createUsers > 0 {
		fmt.Println("Creating users...")
		userScripts.CreateRandomUsers(*createUsers)
	}

	// Delete data if the flag is set
	if *deleteData {
		if *confirmDelete || promptForConfirmation() {
			fmt.Println("Deleting data...")
			dbScripts.RemoveDBData()
			dbScripts.RemoveRedisData()
		} else {
			fmt.Println("Data deletion canceled.")
		}
	}

	if *addFriends {
		// Add random friends
		fmt.Println("Adding random friends...")
		userScripts.AddRandomFriends()
	}

	// Create posts
	fmt.Println("Creating posts...")
	postScripts.CreatePostsForUsers(10) // Change the number to the desired number of posts per user
	filename = "posts.txt"
	postScripts.DownloadPostsToFile(filename)

	// Create comments
	fmt.Println("Creating comments...")
	commentScripts.CreateCommentsForUsers(7) // Change the number to the desired number of comments per post
	filename = "comments.txt"
	commentScripts.DownloadCommentsToFile(filename)
}

func promptForConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Are you sure you want to delete all data? This action cannot be undone. (Y/n): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func displayHelp() {
	fmt.Println("Usage: userScripts [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -c N      Create N random users")
	fmt.Println("  -d        Delete all data (requires confirmation)")
	fmt.Println("  -y        Skip confirmation prompt when using -d")
	fmt.Println("  -af       Add random friends to users")
	fmt.Println("  -h, -help Display help information")
}
