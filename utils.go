package fivesimgo

import (
	"fmt"
	"net/url"

	http "github.com/zMrKrabz/fhttp"
)

func (c *client) makeGetRequest(url string, queryValues *url.Values) (*http.Response, error) {
	// Craft the header
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.APIKey),
	}

	// Creates a request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Incapsulate header elements into the request
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// Encode the query values (if any)
	req.URL.RawQuery = queryValues.Encode()

	return c.httpClient.Do(req)
}
