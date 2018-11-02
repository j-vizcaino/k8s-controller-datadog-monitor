package datadog_client

import (
	"context"
)

type Result map[string]interface{}

type Client interface {
	Host() string
	Post(ctx context.Context, resourcePath string, data string) (Result, error)
}
