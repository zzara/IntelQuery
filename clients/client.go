package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// client struct
type Client struct {
	apiKey      string
	baseUrl     string
	searchPath  string
	httpClient  *http.Client
	queryString string
}

// process json
func ProcessResponse(resp string) []byte {
	var result map[string]interface{}
	json.Unmarshal([]byte(resp), &result)

	file, _ := json.MarshalIndent(result, "", " ")
	return file
}

// client query function
func (s *Client) Query(query string) []byte {

	url, err := url.Parse(s.queryString + query)

	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("GET", url.String(), nil)

	if err != nil {
		fmt.Println(err)
	}
	resp, err := s.httpClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	respBodyStr := string(body)

	file := ProcessResponse(respBodyStr)
	return file
}
