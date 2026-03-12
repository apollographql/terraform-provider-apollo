// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type graphQLRequest struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

type graphQLEnvelope struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

type graphQLError struct {
	Message    string                 `json:"message"`
	Path       []any                  `json:"path,omitempty"`
	Extensions map[string]any         `json:"extensions,omitempty"`
	Locations  []graphQLErrorLocation `json:"locations,omitempty"`
}

type graphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (e graphQLError) code() string {
	if e.Extensions == nil {
		return ""
	}

	value, ok := e.Extensions["code"]
	if !ok {
		value, ok = e.Extensions["errorCode"]
	}

	if !ok {
		return ""
	}

	return strings.ToUpper(fmt.Sprint(value))
}

func (e graphQLError) hasCode(codes ...string) bool {
	actual := e.code()
	if actual == "" {
		return false
	}

	for _, code := range codes {
		if actual == strings.ToUpper(code) {
			return true
		}
	}

	return false
}

func (e graphQLError) statusCode() int {
	for _, key := range []string{"status", "httpStatus"} {
		if value, ok := e.Extensions[key]; ok {
			if status := coerceStatusCode(value); status != 0 {
				return status
			}
		}
	}

	if response, ok := e.Extensions["response"].(map[string]any); ok {
		if status := coerceStatusCode(response["status"]); status != 0 {
			return status
		}
	}

	if httpValue, ok := e.Extensions["http"].(map[string]any); ok {
		if status := coerceStatusCode(httpValue["status"]); status != 0 {
			return status
		}
	}

	return 0
}

func coerceStatusCode(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		number, err := typed.Int64()
		if err == nil {
			return int(number)
		}
	case string:
		number, err := strconv.Atoi(typed)
		if err == nil {
			return number
		}
	}

	return 0
}

func (c *Client) doGraphQL(ctx context.Context, operation string, query string, variables map[string]any, out any) error {
	requestPayload, err := json.Marshal(graphQLRequest{
		Query:         query,
		Variables:     variables,
		OperationName: operation,
	})
	if err != nil {
		return inputError(operation, fmt.Sprintf("unable to encode GraphQL request: %s", err))
	}

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		err = c.doGraphQLAttempt(ctx, operation, requestPayload, out)
		if err == nil {
			return nil
		}

		var apiErr *APIError
		if !AsAPIError(err, &apiErr) || !apiErr.Retryable || attempt == c.maxAttempts {
			return err
		}

		if sleepErr := sleepWithContext(ctx, c.retryDelay(attempt)); sleepErr != nil {
			return sleepErr
		}
	}

	return err
}

func (c *Client) doGraphQLAttempt(ctx context.Context, operation string, requestPayload []byte, out any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(requestPayload))
	if err != nil {
		return classifyOperationError(operation, 0, nil, "unable to create GraphOS Platform API request", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("x-api-key", c.apiKey)
	if c.userAgent != "" {
		request.Header.Set("User-Agent", c.userAgent)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return classifyOperationError(operation, 0, nil, "unable to reach the GraphOS Platform API", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return classifyOperationError(operation, response.StatusCode, nil, "unable to read the GraphOS Platform API response", err)
	}

	var envelope graphQLEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		if response.StatusCode >= 400 {
			return classifyOperationError(operation, response.StatusCode, nil, strings.TrimSpace(string(body)), nil)
		}

		return malformedResponseError(operation, err)
	}

	if response.StatusCode >= 400 {
		statusCode := response.StatusCode
		for _, graphQLError := range envelope.Errors {
			if candidate := graphQLError.statusCode(); candidate != 0 {
				statusCode = candidate
				break
			}
		}

		return classifyOperationError(operation, statusCode, envelope.Errors, response.Status, nil)
	}

	if len(envelope.Errors) > 0 {
		statusCode := 0
		for _, graphQLError := range envelope.Errors {
			if candidate := graphQLError.statusCode(); candidate != 0 {
				statusCode = candidate
				break
			}
		}

		return classifyOperationError(operation, statusCode, envelope.Errors, "", nil)
	}

	if out == nil {
		return nil
	}

	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return malformedResponseError(operation, fmt.Errorf("response did not include a data payload"))
	}

	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return malformedResponseError(operation, err)
	}

	return nil
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func AsAPIError(err error, target **APIError) bool {
	if target == nil {
		return false
	}

	return errorAs(err, target)
}

func errorAs(err error, target **APIError) bool {
	switch typed := err.(type) {
	case *APIError:
		*target = typed
		return true
	case interface{ Unwrap() error }:
		return errorAs(typed.Unwrap(), target)
	default:
		return false
	}
}
