package clients

import (
	"net/http"
)

// create urlscan client
// curl "https://urlscan.io/api/v1/search/?q=domain:urlscan.io&size=1&offset=0"
func UrlscanClient(apiKey string) *Client {
	urlscanBasePath := "https://urlscan.io"
	urlscanSearchPath := "/api/v1/search/"

	return &Client{
		apiKey:      apiKey,
		baseUrl:     urlscanBasePath,
		searchPath:  urlscanSearchPath,
		httpClient:  &http.Client{},
		queryString: urlscanBasePath + urlscanSearchPath + "?q=",
	}
}
