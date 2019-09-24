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

// create shodan client
// curl -X GET 'https://api.shodan.io/shodan/host/search?key<apikey>&query=http.title:paypal&page=0'
func NewShodanClient() *QueryClient {
	basePath := "https://api.shodan.io"
	searchPath := "/shodan/host/search"
	apiKey := os.Getenv("SHODAN_KEY")

	return &QueryClient{
		clientType:  "shodan",
		queries:     queryLoader("shodan_queries_test"),
		apiKey:      apiKey,
		baseUrl:     basePath,
		searchPath:  searchPath,
		queryString: basePath + searchPath + "?key=" + apiKey + "&query=",
		queryPage:   "&page=",
		startPage:   1,
		rateLimit:   1100,
	}
}

// client query function
func shodanQuery(query *Query, pq *[]Query, ch chan string, wg2 *sync.WaitGroup) {

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
			matches = contain["matches"].([]interface{})
			log.Println("message=found_results")
		} else if query.Page > 0 {
			matches = contain["matches"].([]interface{})
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

	// Extract iocs and return any results
	data := extractIOCs(matches)
	wg2.Done()
	for _, d := range data {
		ch <- d
	}
}
