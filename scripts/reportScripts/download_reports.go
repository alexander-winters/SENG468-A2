package reportScripts

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func DownloadReports(username string) {
	urls := []string{
		fmt.Sprintf("http://localhost/reports/%s/posts", username),
		fmt.Sprintf("http://localhost/reports/%s/comments", username),
		fmt.Sprintf("http://localhost/reports/%s/likes", username),
	}

	filenames := []string{
		username + "_posts.json",
		username + "_comments.json",
		username + "_likes.json",
	}

	for i, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching data from %s: %v", url, err)
			continue
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading data from %s: %v", url, err)
			continue
		}

		var jsonData interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			log.Printf("Error unmarshalling JSON data from %s: %v", url, err)
			continue
		}

		prettyJson, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			log.Printf("Error formatting JSON data from %s: %v", url, err)
			continue
		}

		err = os.WriteFile(filenames[i], prettyJson, 0644)
		if err != nil {
			log.Printf("Error writing JSON data to %s: %v", filenames[i], err)
			continue
		}

		fmt.Printf("Data from %s saved to %s\n", url, filenames[i])
	}
}
