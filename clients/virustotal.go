package clients

import (
	"net/http"
)

// create virustotal client
//https://www.virustotal.com/vtapi/v2/url/report?apikey=<apikey>&resource=<resource>
func ShodanClient(apiKey string) *Client {
	shodanBasePath := "https://www.virustotal.com"
	shodanSearchPath := "/vtapi/v2/url/report"

	return &Client{
		apiKey:      apiKey,
		baseUrl:     shodanBasePath,
		searchPath:  shodanSearchPath,
		httpClient:  &http.Client{},
		queryString: shodanBasePath + shodanSearchPath + "?apikey=" + apiKey + "&resource=",
	}
}
