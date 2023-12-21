package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	const query = `{
        "search_type": "match",
        "query":
        {
            "term": "DEMTSCHENKO",
            "start_time": "2021-06-02T14:28:31.894Z",
            "end_time": "2021-12-02T15:28:31.894Z"
        },
        "from": 0,
        "max_results": 20,
        "_source": []
    }`
	req, err := http.NewRequest("POST", "http://localhost:4080/api/olypmics/_search", strings.NewReader(query))
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("admin", "Complexpass#123")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}
