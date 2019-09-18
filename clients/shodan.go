package clients

import (
	"net/http"
)

// create shodan client
func ShodanClient(apiKey string) *Client {
	shodanBasePath := "https://api.shodan.io"
	shodanSearchPath := "/shodan/host/search"

	return &Client{
		apiKey:      apiKey,
		baseUrl:     shodanBasePath,
		searchPath:  shodanSearchPath,
		httpClient:  &http.Client{},
		queryString: shodanBasePath + shodanSearchPath + "?key=" + apiKey + "&query=",
	}
}
