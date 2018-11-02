package datadog_client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type client struct {
	baseURL url.URL
	client  *http.Client
}

const (
	APIContentType = "application/json"
)

func NewClient(hostname string, apiKey string, appKey string) (Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}
	if appKey == "" {
		return nil, fmt.Errorf("APP key cannot be empty")
	}

	u := url.URL{
		Scheme: "https",
		Host: hostname,
	}
	q := u.Query()
	q.Set("api_key", apiKey)
	q.Set("application_key", appKey)
	u.RawQuery = q.Encode()

	return &client{
		baseURL: u,
		client: &http.Client{},
	}, nil
}

func (c *client) resourceURL(resourcePath string) string {
	u := c.baseURL
	u.Path = resourcePath
	return u.String()
}

func (c *client) Host() string {
	return c.baseURL.Host
}

func (c *client) Post(ctx context.Context, resourcePath string, data string) (result Result, err error) {
	r := strings.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, c.resourceURL(resourcePath), r)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", APIContentType)
	req = req.WithContext(ctx)

	res, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("server returned %s", res.Status)
		return
	}

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&result)
	return
}