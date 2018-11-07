package datadog_client

import (
	"context"
)

type Result struct {
	Status int
	Data   map[string]interface{}
}

type Client interface {
	Host() string
	Request(ctx context.Context, method string, resourcePath string, data string) (Result, error)
}
