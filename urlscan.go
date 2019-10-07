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
	"time"
)

// Create urlscan client
// Curl "https://urlscan.io/api/v1/search/?q=domain:urlscan.io&size=1&offset=0"
func NewUrlscanClient() *QueryClient {
	basePath := "https://urlscan.io"
	searchPath := "/api/v1/search/"
	apiKey := os.Getenv("URLSCAN_KEY")

	today := time.Now().Format(time.RFC3339)
	date, _ := time.Parse(time.RFC3339, today)

	return &QueryClient{
		clientType:  "urlscan",
		queries:     queryLoader("urlscan_queries"),
		apiKey:      apiKey,
		baseUrl:     basePath,
		searchPath:  searchPath,
		queryString: basePath + searchPath + "?q=",
		queryPage:   fmt.Sprintf("%%20AND%%20date:%d-%02d-%02d&size=100&offset=", date.Year(), date.Month(), date.Day()),
		startPage:   0,
		rateLimit:   2000,
	}
}

func urlscanParseTime(data []interface{}) []interface{} {
	now := time.Now()
	eightHoursBack := now.Add(time.Duration(-480) * time.Minute)

	layout := "2006-01-02T15:04:05.000Z"

	var filteredData []interface{}

	for _, d := range data {

		resultTimeString := fmt.Sprintf("%v", d.(map[string]interface{})["task"].(map[string]interface{})["time"])
		resultTime, err := time.Parse(layout, resultTimeString)

		if err != nil {
			fmt.Println(err)
		}

		if resultTime.After(eightHoursBack) {
			//fmt.Println(resultTime)
			filteredData = append(filteredData, d)
		}
	}
	return filteredData
}

// Client query function
func urlscanQuery(query *Query, pq *[]Query, ch chan string, wg2 *sync.WaitGroup) {

	url, err := url.Parse(query.Query + strconv.Itoa(query.Page))

	if err != nil {
		fmt.Println(err)
	}

	// Handle http request and response
	respBodyStr := handleRequest(url)
	file := processResponse(respBodyStr)

	// A map container to decode the JSON structure into
	contain := make(map[string]interface{})

	// Unmarshal JSON
	e := json.Unmarshal(file, &contain)

	// Panic on error
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
				*pq = append(*pq, Query{Query: query.Query, Page: i * 100})
				//fmt.Println(i)
			}
		}
	} else {
		log.Println("message=no_results_for_query error=results_nil")
	}

	timeFiltered := urlscanParseTime(matches)
	data := extractIOCs(timeFiltered)

	for i := 0; i < len(data); i++ {
		if i+2 > len(data) {
			break
		} else {
			go func(ch chan<- string, data string) {
				ch <- data
			}(ch, data[i])
		}
	}
	wg2.Done()
}
