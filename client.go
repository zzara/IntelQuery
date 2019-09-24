package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Json map for loading queries from the Json query file
type QueryMapper struct {
	Name  string
	Query string
}

// Struct for holding all of the API relevant information and parameters
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

// Struct for maintaining the query syntax and page number
type Query struct {
	Query string
	Page  int
}

// Request func for iterating through API calls to different services
func requestor(qc *QueryClient, clientQuery func(*Query, *[]Query, chan string, *sync.WaitGroup), ch chan string, wg *sync.WaitGroup) {
	// Wait group for queryApiChannels
	//defer wg.Done()

	// Establish waitgroup for clientQuery requests
	var wg2 sync.WaitGroup

	var pagequeries []Query

	// Iterate queries loaded from Json, pass them to the clientQuery function
	for _, ci := range qc.queries {
		wg2.Add(1)
		query := &Query{Query: qc.queryString + ci.Query + qc.queryPage, Page: qc.startPage}
		go clientQuery(query, &pagequeries, ch, &wg2)
		time.Sleep(time.Millisecond * time.Duration(qc.rateLimit))
	}
	wg2.Wait()

	// Iterate through pages of queries
	if len(pagequeries) > 0 {
		for _, ci := range pagequeries {
			wg2.Add(1)
			query := &Query{Query: qc.queryString + ci.Query + qc.queryPage, Page: ci.Page}
			go clientQuery(query, &pagequeries, ch, &wg2)
			time.Sleep(time.Millisecond * time.Duration(qc.rateLimit))
		}
	}

	// Wait untl all of the requests have been processed before closing the channel
	wg2.Wait()
	wg.Done()
	return
}

// Use regexp to look for domains and URLs in the matched queries
func extractIOCs(data []interface{}) []string {
	var resultIOCs []string
	urlRe := regexp.MustCompile("((?:ht)?f?tps?:\\/\\/?(?:[a-z0-9_]{1,}\\.){1,}[a-z]{1,10}(?:\\/(?:[a-zA-Z0-9_\\-\\?\\!\\@\\#\\:\\(\\)\\=\\+\\&]{1,})?){0,})")
	ipRe := regexp.MustCompile("([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})")
	for _, d := range data {
		str := fmt.Sprintf("%v", d)
		//fmt.Println(str)
		urlMatch := urlRe.FindStringSubmatch(str)
		if urlMatch != nil {
			for _, ioc := range urlMatch[1:] {
				fmt.Println(ioc)
				resultIOCs = append(resultIOCs, fmt.Sprintf("%s", ioc))
			}
		}
		ipMatch := ipRe.FindStringSubmatch(str)
		if ipMatch != nil {
			for _, ioc := range ipMatch[1:] {
				fmt.Println(ioc)
				resultIOCs = append(resultIOCs, fmt.Sprintf("%s", ioc))
			}
		}
	}
	return resultIOCs
}

// Api streams, main function for initializing clients
func queryApiChannels() {
	// Create api clients and load json queries
	ch := make(chan string)
	var wg sync.WaitGroup

	wg.Add(1)
	go requestor(NewShodanClient(), shodanQuery, ch, &wg)
	wg.Add(1)
	go requestor(NewUrlscanClient(), urlscanQuery, ch, &wg)

	wg.Wait()

	// Close channel and get data
	close(ch)
	data := <-ch

	// Print data from channel
	for _, d := range data {
		fmt.Println(d)
	}
}

// Processes Json response into []byte format
func processResponse(resp string) []byte {
	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	file, _ := json.MarshalIndent(result, "", " ")
	return file
}

// Handle the request to the API
func handleRequest(url *url.URL) string {
	httpClient := &http.Client{Timeout: time.Second * 10}

	req, err := http.NewRequest("GET", url.String(), nil)

	fmt.Println(url.String())

	if err != nil {
		fmt.Println(err)
	}
	resp, err := httpClient.Do(req)

	var respBodyStr string
	var body []byte

	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
	}

	respBodyStr = string(body)
	return respBodyStr
}

// Loads and maps each query from the Json query file
func queryLoader(filename string) []QueryMapper {

	// Get folder path of the current binary execution
	_, goBinPath, _, _ := runtime.Caller(0)
	// Split file path
	dirSplitPath := strings.Split(goBinPath, "/")
	// Remove last element of array
	dirPathMod := dirSplitPath[:len(dirSplitPath)-1]
	// join path
	dirPath := strings.Join(dirPathMod, "/")

	// Load the Json query file into f
	f, err := os.Open(fmt.Sprintf("%s/queries/%s%s", dirPath, filename, ".json"))
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
