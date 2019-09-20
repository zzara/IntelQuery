package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
)

// create shodan client
func NewShodanClient() *QueryClient {
	basePath := "https://api.shodan.io"
	searchPath := "/shodan/host/search"
	apiKey := os.Getenv("SHODAN_KEY")

	return &QueryClient{
		clientType:  "shodan",
		queries:     queryLoader("shodan_queries"),
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
func shodanQuery(query *Query, pq *[]Query, chnl chan []interface{}, wg2 *sync.WaitGroup) {
	defer wg2.Done()

	httpClient := &http.Client{}

	url, err := url.Parse(query.Query + strconv.Itoa(query.Page))

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("GET", url.String(), nil)

	fmt.Println(url.String())

	if err != nil {
		fmt.Println(err)
	}
	resp, err := httpClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	respBodyStr := string(body)

	fmt.Println(len(respBodyStr))

	var result map[string]interface{}
	json.Unmarshal([]byte(respBodyStr), &result)

	file, _ := json.MarshalIndent(result, "", " ")

	// a map container to decode the JSON structure into
	contain := make(map[string]interface{})

	// unmarshal JSON
	e := json.Unmarshal(file, &contain)

	// panic on error
	if e != nil {
		panic(e)
	}

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
				*pq = append(*pq, Query{Query: query.Query, Page: i})
				fmt.Println(i)
			}
		}
	} else {
		log.Println("message=no_results_for_query error=results_nil")
	}

	chnl <- matches
}
