package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alexander-winters/SENG468-A2/scripts/commentScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/dbScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/likeScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/postScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/reportScripts"
	"github.com/alexander-winters/SENG468-A2/scripts/userScripts"
)

func main() {
	// Define the command-line flags
	createUsers := flag.Int("c", 0, "Number of users to create")
	createPosts := flag.Int("cp", 0, "Number of posts to create per user")
	createComments := flag.Int("cc", 0, "Number of comments to create per user per post")
	deleteData := flag.Bool("d", false, "Delete all data (requires confirmation)")
	confirmDelete := flag.Bool("y", false, "Skip confirmation prompt when using -d")
	addFriends := flag.Bool("af", false, "Add random friends to users")
	userReports := flag.String("r", "", "Download user reports for specified username")
	generateLikes := flag.Bool("g", false, "Randomly like posts and comments")
	help := flag.Bool("h", false, "Display help information")
	helpLong := flag.Bool("help", false, "Display help information")

	// Parse the command-line flags
	flag.Parse()

	if *help || *helpLong {
		displayHelp()
		return
	}

	// Delete data if the flag is set
	if *deleteData {
		if *confirmDelete || promptForConfirmation() {
			fmt.Println("Deleting data...")
			dbScripts.RemoveDBData()
			dbScripts.RemoveRedisData()
			fmt.Println("Data deletion completed. Exiting.")
			os.Exit(0) // Exit the program after data deletion
		} else {
			fmt.Println("Data deletion canceled.")
		}
	}

	// Create users if the flag is set
	if *createUsers > 0 {
		fmt.Println("Creating users...")
		userScripts.CreateRandomUsers(*createUsers)
		// Download users
		filename := "users.txt"
		fmt.Println("Downloading users...")
		userScripts.DownloadUsersToFile(filename)
	}

	if *addFriends {
		// Add random friends
		fmt.Println("Adding random friends...")
		userScripts.AddRandomFriends()
	}

	// Create posts if the flag is set
	if *createPosts > 0 {
		// Create posts
		fmt.Println("Creating posts...")
		postScripts.CreatePostsForUsers(*createPosts) // Change the number to the desired number of posts per user
		filename := "posts.txt"
		postScripts.DownloadPostsToFile(filename)
	}

	if *createComments > 0 {
		// Create comments
		fmt.Println("Creating comments...")
		commentScripts.CreateCommentsForUsers(*createComments) // Change the number to the desired number of comments per post
		filename := "comments.txt"
		commentScripts.DownloadCommentsToFile(filename)
	}

	if *userReports != "" {
		// Get user reports
		username := *userReports
		fmt.Println("Getting user reports")
		reportScripts.DownloadReports(username)
	}

	if *generateLikes {
		// Generate random likes on posts and comments
		first_file := "posts.txt"
		fmt.Println("Generating likes from both files...")
		second_file := "comments.txt"
		likeScripts.LikePostsAndComments(first_file, second_file)
	}
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
	fmt.Println("Usage: scripts [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -c N          Create N random users")
	fmt.Println("  -cp N         Create N random posts per user")
	fmt.Println("  -cc N         Create N random comments per user per post")
	fmt.Println("  -d            Delete all data (requires confirmation)")
	fmt.Println("  -y            Skip confirmation prompt when using -d")
	fmt.Println("  -af           Add random friends to users")
	fmt.Println("  -r <username> Get reports for specified username")
	fmt.Println("  -g            Generate likes for posts and comments randomly")
	fmt.Println("  -h, -help     Display help information")
}
