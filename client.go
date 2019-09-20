package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// Json map for loading queries from the Json query file
type QueryMapper struct {
	Name  string
	Query string
}

type QueryClient struct {
	clientType  string
	queries     []QueryMapper
	apiKey      string
	baseUrl     string
	searchPath  string
	queryString string
	queryPage   string
	startPage   int
	rateLimit   int
}

type Query struct {
	Query string
	Page  int
}

// In the future, type will be: chnl chan []interface{}
// This will be due to the map iterator for finding things

// Request func for iterating through API calls to different services
func requestor(s *QueryClient, clientQuery func(*Query, *[]Query, chan []interface{}, *sync.WaitGroup), chnl chan []interface{}, wg *sync.WaitGroup) {
	// Wait group for queryApiChannels
	defer wg.Done()

	// Establish waitgroup for clientQuery requests
	var wg2 sync.WaitGroup
	var pagequeries []Query

	// Iterate queries loaded from Json, pass them to the clientQuery function
	for _, ci := range s.queries {
		wg2.Add(1)
		query := &Query{Query: s.queryString + ci.Query + s.queryPage, Page: s.startPage}
		go clientQuery(query, &pagequeries, chnl, &wg2)
		time.Sleep(time.Millisecond * time.Duration(s.rateLimit))
	}
	wg2.Wait()
	fmt.Println("YEES")
	// Iterate through pages of queries
	if len(pagequeries) > 0 {
		for _, ci := range pagequeries {
			wg2.Add(1)
			query := &Query{Query: s.queryString + ci.Query + s.queryPage, Page: ci.Page}
			go clientQuery(query, &pagequeries, chnl, &wg2)
			time.Sleep(time.Millisecond * time.Duration(s.rateLimit))
		}
	}

	// Wait untl all of the requests have been processed before closing the channel
	wg2.Wait()
}

// api streams, main function for initializing clients
func queryApiChannels() {
	// Create api clients and load json queries
	ch := make(chan []interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go requestor(NewShodanClient(), shodanQuery, ch, &wg)
	wg.Add(1)
	go requestor(NewUrlscanClient(), urlscanQuery, ch, &wg)
	wg.Wait()
	close(ch)

	for {
		v, ok := <-ch
		if ok == false {
			break
		}
		fmt.Println("Received ", len(v), ok)
	}
	wg.Wait()
}

// Processes Json response into []byte format
func ProcessResponse(resp string) []byte {
	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	file, _ := json.MarshalIndent(result, "", " ")
	return file
}

// Loads and maps each query from the Json query file
func queryLoader(filename string) []QueryMapper {

	// Get folder pth of the current binary execution
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Load the Json query file into f
	f, err := os.Open(fmt.Sprintf("%s/%s%s", dir, filename, ".json"))
	if err != nil {
		log.Printf("message=failed_to_open_query_file error=%s\n", err)
	}

	log.Printf("status=successfully_opened_query_json_file message=%s\n", filename)
	defer f.Close()

	// Load and map the Json query file
	byteValue, _ := ioutil.ReadAll(f)
	var result []QueryMapper
	json.Unmarshal(byteValue, &result)

	return result
}
