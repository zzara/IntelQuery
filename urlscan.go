package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"strconv"
	"sync"
)

// create urlscan client
// curl "https://urlscan.io/api/v1/search/?q=domain:urlscan.io&size=1&offset=0"
func NewUrlscanClient() *QueryClient {
	basePath := "https://urlscan.io"
	searchPath := "/api/v1/search/"
	apiKey := os.Getenv("URLSCAN_KEY")

	return &QueryClient{
		clientType:  "urlscan",
		queries:     queryLoader("urlscan_queries_test"),
		apiKey:      apiKey,
		baseUrl:     basePath,
		searchPath:  searchPath,
		queryString: basePath + searchPath + "?q=",
		queryPage:   "&offset=",
		startPage:   0,
		rateLimit:   2000,
	}
}

// client query function
func urlscanQuery(query *Query, pq *[]Query, ch chan string, wg2 *sync.WaitGroup) {

	url, err := url.Parse(query.Query + strconv.Itoa(query.Page))

	if err != nil {
		fmt.Println(err)
	}

	// Handle http request and response
	respBodyStr := handleRequest(url)
	file := processResponse(respBodyStr)

	// a map container to decode the JSON structure into
	contain := make(map[string]interface{})

	// unmarshal JSON
	e := json.Unmarshal(file, &contain)

	// panic on error
	if e != nil {
		panic(e)
	}

	// Look for matches in the http response data
	var matches []interface{}
	var totals float64
	if contain["total"] != nil {
		totals = contain["total"].(float64)

		if totals == float64(0) {
			log.Println("message=no_results error=number_of_results_is_0")
		} else if totals <= float64(100) {
			matches = contain["results"].([]interface{})
			log.Println("message=found_results")
		} else if query.Page > 0 {
			matches = contain["results"].([]interface{})
			log.Println("message=found_results")
		} else {
			pages := int(math.Ceil(totals / 100))
			var i int
			for i = 1; i <= pages; i++ {
				// Create a new page query for any total over 100 results
				*pq = append(*pq, Query{Query: query.Query, Page: i})
				fmt.Println(i)
			}
		}
	} else {
		log.Println("message=no_results_for_query error=results_nil")
	}

	data := extractIOCs(matches)
	wg2.Done()
	for _, d := range data {
		ch <- d
	}
}
