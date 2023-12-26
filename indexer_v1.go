package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Email represents the structure of the email data.
type Email struct {
	ID                        int    `json:"ID"`
	Message_ID                string `json:"Message-ID"`
	Date                      string `json:"Date"`
	From                      string `json:"from"`
	To                        string `json:"to"`
	Subject                   string `json:"subject"`
	Mime_Version              string `json:"Mime-Version"`
	Content_Type              string `json:"Content-Type"`
	Content_Transfer_Encoding string `json:"Content-Transfer-Encoding"`
	X_From                    string `json:"X-From"`
	X_To                      string `json:"X-To"`
	X_cc                      string `json:"X-cc"`
	X_bcc                     string `json:"X-bcc"`
	X_Folder                  string `json:"X-Folder"`
	X_Origin                  string `json:"X-Origin"`
	X_FileName                string `json:"X-FileName"`
	Cc                        string `json:"Cc"`
	Body                      string `json:"Body"`
}

// Constants for configuration.
const (
	ZincHost    = "http://localhost:4080"
	Index       = "enronJELM"
	Credentials = "admin:Complexpass#123"
)

func main() {
	// Setup logging
	logFile, err := os.Create("app.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	// Perform indexing
	err = indexEmails("/Users/jorgecapcha/Documents/GO/enron_mail_20110402/maildir")
	if err != nil {
		log.Fatalf("Indexing failed: %v", err)
	}
}

func indexEmails(rootPath string) error {
	users, err := listFolders(rootPath)
	if err != nil {
		return err
	}

	// Create a buffer to accumulate NDJSON data
	var ndjsonBuffer bytes.Buffer

	for _, user := range users {
		folders, err := listFolders(filepath.Join(rootPath, user))
		if err != nil {
			return err
		}

		for _, folder := range folders {
			files, err := listFiles(filepath.Join(rootPath, user, folder))
			if err != nil {
				return err
			}

			for _, file := range files {
				email, err := processEmail(filepath.Join(rootPath, user, folder, file))
				if err != nil {
					log.Printf("Error processing email %s: %v", file, err)
					continue
				}

				// Convert each email to NDJSON format and append to the buffer
				ndjsonLine, err := json.Marshal(email)
				if err != nil {
					log.Printf("Error encoding email to JSON: %v", err)
					continue
				}
				ndjsonBuffer.WriteString(string(ndjsonLine) + "\n")
			}
		}
	}

	// Now, ndjsonBuffer contains all emails in NDJSON format

	// Perform the indexing using NDJSON data
	err = indexNDJSON(ndjsonBuffer.String())
	if err != nil {
		log.Printf("Error indexing emails: %v", err)
		return err
	}

	return nil
}

func indexNDJSON(ndjsonData string) error {
	user := "admin"
	password := "Complexpass#123"
	auth := user + ":" + password
	bas64encodedCreds := base64.StdEncoding.EncodeToString([]byte(auth))
	index := "enronJELM"
	zincHost := "http://localhost:4080"
	zincURL := zincHost + "/api/" + index + "/_bulk"

	// Create a new HTTP request
	req, err := http.NewRequest("POST", zincURL, strings.NewReader(ndjsonData))
	if err != nil {
		log.Fatal("Error creating request. ", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-ndjson")
	req.Header.Set("Authorization", "Basic "+bas64encodedCreds)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	// Send the request to Zincsearch
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Non-OK status code: %d", resp.StatusCode)
	}

	return nil
}

func listFolders(folderPath string) ([]string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var folders []string
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file.Name())
		}
	}
	return folders, nil
}

func listFiles(folderPath string) ([]string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames, nil
}

func processEmail(filePath string) (*Email, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the entire email content into a string
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Split the content into headers and body
	headerBodySplit := strings.SplitN(string(content), "\n\n", 2)
	if len(headerBodySplit) < 2 {
		return nil, fmt.Errorf("malformed email, missing body")
	}

	headers := headerBodySplit[0]
	body := headerBodySplit[1]

	// Parse headers
	scanner := bufio.NewScanner(strings.NewReader(headers))
	emailData := &Email{}
	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.SplitN(line, ":", 2)
		if len(fields) == 2 {
			header, value := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
			switch header {
			case "Message-ID":
				emailData.Message_ID = value
			case "Date":
				emailData.Date = value
			case "From":
				emailData.From = value
			case "To":
				emailData.To = value
			case "Subject":
				emailData.Subject = value
			case "Cc":
				emailData.Cc = value
			case "Mime-Version":
				emailData.Mime_Version = value
			case "Content-Type":
				emailData.Content_Type = value
			case "Content-Transfer-Encoding":
				emailData.Content_Transfer_Encoding = value
			case "X-From":
				emailData.X_From = value
			case "X-To":
				emailData.X_To = value
			case "X-cc":
				emailData.X_cc = value
			case "X-bcc":
				emailData.X_bcc = value
			case "X-Folder":
				emailData.X_Folder = value
			case "X-Origin":
				emailData.X_Origin = value
			case "X-FileName":
				emailData.X_FileName = value
			}
		}
	}

	// Set the email body
	emailData.Body = body

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return emailData, nil
}
