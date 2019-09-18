package clients

import (
	"net/http"
)

// create urlscan client
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
