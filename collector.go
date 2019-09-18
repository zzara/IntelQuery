package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"intelquery/clients"
	"runtime"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

// logging
type LogWriter struct {
}

func (writer LogWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + " [DEBUG] " + string(bytes))
}

// constants
const GLOBAL_LIMITER time.Duration = 1100
const OUTPUT_FOLDER string = "/tmp/gonyr/"

// json map for loading json queries
type JsonMapper struct {
	Name  string
	Query string
}

type Client struct {
	Type    string
	Queries []JsonMapper
	Client  interface{}
}

// lambda handler
type MyEvent struct {
	Name string `json:"name"`
}

// load client return json
func jsonLoader(filename string) []JsonMapper {

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(fmt.Sprintf("%s/%s%s", dir, filename, ".json"))
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("status=successfully_opened_query_json_file message=%s\n", filename)
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)
	var result []JsonMapper
	json.Unmarshal(byteValue, &result)

	return result
}

// write out the returned content
func fileHandler(file []byte, out string, id string) {
	location := fmt.Sprintf("%s%s.json", out, id)
	_ = ioutil.WriteFile(location, file, 0644)
	return
}

// process and parse json responses
func extractIocData(file []byte) {
	// a map container to decode the JSON structure into
	c := make(map[string]interface{})

	// unmarshal JSON
	e := json.Unmarshal(file, &c)

	// panic on error
	if e != nil {
		panic(e)
	}

	// a string slice to hold the keys
	k := make([]string, len(c))

	totals := c["total"].(float64)

	if totals <= float64(100) {
		if c["matches"] != nil {

			//matches := c["matches"].([]interface{})
			fmt.Println("yup")

			/*
				if totals <= 100 {
					for key, val := range matches {
						fmt.Println(key, val)
						subvals := val.(map[string]interface{})
						for y, _ := range subvals {
							fmt.Println(y)
						}
					}
				} else {
					fmt.Println("total greater than 100")
				}
			*/
		}
	} else {
		pages := math.Ceil(totals / 100)
		fmt.Println(pages)
	}

	// iteration counter
	i := 0

	// copy c's keys into k
	for s, _ := range c {
		k[i] = s
		i++
	}

	for i, _ := range k {
		fmt.Println(k[i])
	}

	return
}

// init queries for each query client
func queryApi(client *Client, clientQueryFunction func(string) []byte, apiClientChannel chan int, clientWaitGroup *sync.WaitGroup) {
	// init clientWaitGroup clients wait group
	defer clientWaitGroup.Done()

	// create query channels
	queryClientChannel := make(chan int, len(client.Queries))
	// iterate through queries to create channels
	for n := range client.Queries {
		queryClientChannel <- n
	}
	close(queryClientChannel)

	// limiter for api query time
	limiter := time.Tick(time.Millisecond * GLOBAL_LIMITER)

	// create wait group for lambda processes, length of number of queries
	var queryWaitGroup sync.WaitGroup
	queryWaitGroup.Add(len(client.Queries))

	// iterate, one query for each channel
	for n := range queryClientChannel {
		<-limiter // wait time specified by GLOBAL_LIMITER variable

		// lambda function to query and process each query and result
		go func(n int) {
			// remove one task from the wait group
			defer queryWaitGroup.Done()

			// set name and query parameters
			name := client.Type + "_" + client.Queries[n].Name
			query := client.Queries[n].Query

			log.Printf("status=querying message=id:%s,query:%s\n", name, query)

			// build query response
			var file []byte
			file = clientQueryFunction(query) // initiate query

			if len(file) > 20 {
				// parse and process json result for iocs
				go extractIocData(file)

				// handle output file, write to disk
				//fmt.Printf("Results found. Saving results to %s\n", OUTPUT_FOLDER)
				//fileHandler(file, OUTPUT_FOLDER, name)
				//fmt.Printf("%s\n", file)
			} else if len(file) > 20 {
				log.Println("status=failed message=invalid_return_data_length")
			}
		}(n)
	}

	result := 1
	log.Printf("status=success message=%s_client_operation_completed\n", client.Type)

	// wait for wait group to finish
	queryWaitGroup.Wait()

	apiClientChannel <- result
}

// api streams, main function for initializing clients
func openApiChannels() {
	// create api clients and load json queries
	sc := &Client{Type: "shodan", Queries: jsonLoader("shodan_queries"), Client: clients.ShodanClient(os.Getenv("SHODAN_KEY"))}
	uc := &Client{Type: "urlscan", Queries: jsonLoader("urlscan_queries"), Client: clients.UrlscanClient(os.Getenv("URLSCAN_KEY"))}

	// wait group for clients
	var clientWaitGroup sync.WaitGroup
	// create api client channels
	apiClientChannel := make(chan int, 2)

	// execute api client channel
	clientWaitGroup.Add(1)
	go queryApi(sc, sc.Client.(*clients.Client).Query, apiClientChannel, &clientWaitGroup)
	clientWaitGroup.Add(1)
	go queryApi(uc, uc.Client.(*clients.Client).Query, apiClientChannel, &clientWaitGroup)

	//wait for client wait group to then close the channel
	clientWaitGroup.Wait()

	//close client channel
	close(apiClientChannel)

	for n := range apiClientChannel {
		fmt.Println(n) // value check
	}

	return
}

// lambda handler
func handleRequest(ctx context.Context, name MyEvent) (string, error) {
	openApiChannels()
	return fmt.Sprintf("Hello %s!", name.Name), nil
}

func main() {
	// logging
	log.SetFlags(0)
	log.SetOutput(new(LogWriter))
	log.Println("status=collection_process_started message=started")

	// auto-detect environment for osx, launch non-lambda, for testing
	log.Printf("status=os_detected message=%s\n", runtime.GOOS)
	if runtime.GOOS == "darwin" {
		openApiChannels()
		// auto-detect environment for linux (aws lambda), launch lambda
	} else if runtime.GOOS == "linux" {
		lambda.Start(handleRequest)
	}
}
