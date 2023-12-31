package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	"io"
	"log"

	"net/http"
	"os"
	"path/filepath"
	// "runtime"
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

func main1() {
	// Setup logging
	logFile, err := os.Create("app.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	// TODO: add setup code, if needed

	// Perform indexing
	err = indexEmails("/Users/jorgecapcha/Documents/GO/enron_mail_20110402/maildir")
	if err != nil {
		log.Fatalf("Indexing failed: %v", err)
	}

	// TODO: add cleanup code, if needed
}

func indexEmails(rootPath string) error {
	users, err := listFolders(rootPath)
	if err != nil {
		return err
	}

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

				err = indexData(email)
				if err != nil {
					log.Printf("Error indexing email %s: %v", file, err)
					// TODO: Handle indexing error, if needed
				}
			}

		}
	}

	// TODO: further cleanup or finalization, if

	return nil
}

func indexData(data *Email) error {
	user := "admin"
	password := "Complexpass#123"
	auth := user + ":" + password
	bas64encoded_creds := base64.StdEncoding.EncodeToString([]byte(auth))
	index := "enronJELM"
	zinc_host := "http://localhost:4080"
	zinc_url := zinc_host + "/api/" + index + "/_doc"
	jsonData, _ := json.MarshalIndent(data, "", "   ")
	req, err := http.NewRequest("POST", zinc_url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+bas64encoded_creds)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fmt.println(Body)
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

	dataLines := bufio.NewScanner(file)
	emailData := &Email{}

	for dataLines.Scan() {
		line := dataLines.Text()

		// Parse each line and populate the Email struct
		if strings.Contains(line, "Message-ID:") {
			emailData.Message_ID = line[11:]
		} else if strings.Contains(line, "Date:") {
			emailData.Date = line[5:]
		} else if strings.Contains(line, "From:") {
			emailData.From = line[6:]
		} else if strings.Contains(line, "To:") {
			emailData.To = line[4:]
		} else if strings.Contains(line, "Subject:") {
			emailData.Subject = line[8:]
		} else if strings.Contains(line, "Cc:") {
			emailData.Cc = line[3:]
		} else if strings.Contains(line, "Mime-Version:") {
			emailData.Mime_Version = line[9:]
		} else if strings.Contains(line, "Content-Type:") {
			emailData.Content_Type = line[9:]
		} else if strings.Contains(line, "Content-Transfer-Encoding:") {
			emailData.Content_Transfer_Encoding = line[9:]
		} else if strings.Contains(line, "X-From:") {
			emailData.X_From = line[9:]
		} else if strings.Contains(line, "X-To:") {
			emailData.X_To = line[9:]
		} else if strings.Contains(line, "X-cc:") {
			emailData.X_cc = line[6:]
		} else if strings.Contains(line, "X-bcc:") {
			emailData.X_bcc = line[6:]
		} else if strings.Contains(line, "X-Folder:") {
			emailData.X_Folder = line[9:]
		} else if strings.Contains(line, "X-Origin:") {
			emailData.X_Origin = line[9:]
		} else if strings.Contains(line, "X-FileName:") {
			emailData.X_FileName = line[9:]
		} else {
			emailData.Body += line
		}
	}

	if err := dataLines.Err(); err != nil {
		return nil, err
	}

	return emailData, nil
}
