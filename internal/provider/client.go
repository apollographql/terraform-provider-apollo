// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net/http"
	"time"
)

const (
	defaultClientTimeout = 30 * time.Second
	defaultMaxAttempts   = 3
)

type RetryDelayFunc func(attempt int) time.Duration

type ClientConfig struct {
	APIKey      string
	Endpoint    string
	UserAgent   string
	HTTPClient  *http.Client
	MaxAttempts int
	RetryDelay  RetryDelayFunc
}

type Client struct {
	apiKey      string
	endpoint    string
	httpClient  *http.Client
	userAgent   string
	maxAttempts int
	retryDelay  RetryDelayFunc
}

func NewClient(apiKey string, endpoint string, userAgent string) *Client {
	return NewClientWithConfig(ClientConfig{
		APIKey:    apiKey,
		Endpoint:  endpoint,
		UserAgent: userAgent,
	})
}

func NewClientWithConfig(config ClientConfig) *Client {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultClientTimeout}
	} else {
		cloned := *httpClient
		if cloned.Timeout == 0 {
			cloned.Timeout = defaultClientTimeout
		}

		httpClient = &cloned
	}

	maxAttempts := config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}

	retryDelay := config.RetryDelay
	if retryDelay == nil {
		retryDelay = func(attempt int) time.Duration {
			return time.Duration(attempt) * 100 * time.Millisecond
		}
	}

	return &Client{
		apiKey:      config.APIKey,
		endpoint:    config.Endpoint,
		httpClient:  httpClient,
		userAgent:   config.UserAgent,
		maxAttempts: maxAttempts,
		retryDelay:  retryDelay,
	}
}

func (c *Client) Endpoint() string {
	return c.endpoint
}
