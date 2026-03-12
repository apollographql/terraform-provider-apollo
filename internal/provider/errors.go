// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorKind string

const (
	ErrorKindUnknown            ErrorKind = "unknown"
	ErrorKindUnauthenticated    ErrorKind = "unauthenticated"
	ErrorKindPermissionDenied   ErrorKind = "permission_denied"
	ErrorKindNotFound           ErrorKind = "not_found"
	ErrorKindRateLimited        ErrorKind = "rate_limited"
	ErrorKindTransient          ErrorKind = "transient"
	ErrorKindInvalidRequest     ErrorKind = "invalid_request"
	ErrorKindMalformedResponse  ErrorKind = "malformed_response"
	ErrorKindCapabilityDeferred ErrorKind = "capability_deferred"
)

type APIError struct {
	Kind          ErrorKind
	Operation     string
	HTTPStatus    int
	Message       string
	GraphQLErrors []graphQLError
	Retryable     bool
	Cause         error
}

func (e *APIError) Error() string {
	var details []string

	if e.Operation != "" {
		details = append(details, fmt.Sprintf("operation=%s", e.Operation))
	}

	if e.Kind != "" {
		details = append(details, fmt.Sprintf("kind=%s", e.Kind))
	}

	if e.HTTPStatus != 0 {
		details = append(details, fmt.Sprintf("status=%d", e.HTTPStatus))
	}

	message := e.Message
	if message == "" && e.Cause != nil {
		message = e.Cause.Error()
	}

	if len(details) == 0 {
		return message
	}

	if message == "" {
		return strings.Join(details, " ")
	}

	return fmt.Sprintf("%s (%s)", message, strings.Join(details, ", "))
}

func (e *APIError) Unwrap() error {
	return e.Cause
}

func (e *APIError) Temporary() bool {
	return e.Retryable
}

func IsErrorKind(err error, kind ErrorKind) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.Kind == kind
}

func classifyOperationError(operation string, statusCode int, graphQLErrors []graphQLError, fallbackMessage string, cause error) *APIError {
	kind := ErrorKindUnknown
	retryable := false
	message := fallbackMessage

	if len(graphQLErrors) > 0 {
		message = joinGraphQLErrorMessages(graphQLErrors)
	}

	switch {
	case statusCode == 401 || hasGraphQLErrorCode(graphQLErrors, "UNAUTHENTICATED", "UNAUTHORIZED"):
		kind = ErrorKindUnauthenticated
	case statusCode == 403 || hasGraphQLErrorCode(graphQLErrors, "FORBIDDEN", "PERMISSION_DENIED", "NOT_ALLOWED_BY_USER_ROLE") || messageContains(message, "NOT_ALLOWED_BY_USER_ROLE") || messageContains(message, "not allowed by user role"):
		kind = ErrorKindPermissionDenied
	case statusCode == 404 || hasGraphQLErrorCode(graphQLErrors, "NOT_FOUND") || messageContains(message, "not found"):
		kind = ErrorKindNotFound
	case statusCode == 429 || hasGraphQLErrorCode(graphQLErrors, "RATE_LIMITED", "TOO_MANY_REQUESTS"):
		kind = ErrorKindRateLimited
		retryable = true
	case statusCode >= 500:
		kind = ErrorKindTransient
		retryable = true
	case hasGraphQLErrorCode(graphQLErrors, "GRAPHQL_VALIDATION_FAILED", "BAD_USER_INPUT"):
		kind = ErrorKindInvalidRequest
	case statusCode >= 400:
		kind = ErrorKindInvalidRequest
	}

	if cause != nil && kind == ErrorKindUnknown {
		kind = ErrorKindTransient
		retryable = true
	}

	return &APIError{
		Kind:          kind,
		Operation:     operation,
		HTTPStatus:    statusCode,
		Message:       message,
		GraphQLErrors: graphQLErrors,
		Retryable:     retryable,
		Cause:         cause,
	}
}

func capabilityDeferredError(operation string, message string) *APIError {
	return &APIError{
		Kind:      ErrorKindCapabilityDeferred,
		Operation: operation,
		Message:   message,
	}
}

func malformedResponseError(operation string, cause error) *APIError {
	return &APIError{
		Kind:      ErrorKindMalformedResponse,
		Operation: operation,
		Message:   "the GraphOS Platform API returned a malformed response",
		Cause:     cause,
	}
}

func inputError(operation string, message string) *APIError {
	return &APIError{
		Kind:      ErrorKindInvalidRequest,
		Operation: operation,
		Message:   message,
	}
}

func hasGraphQLErrorCode(errors []graphQLError, codes ...string) bool {
	for _, graphQLError := range errors {
		if graphQLError.hasCode(codes...) {
			return true
		}
	}

	return false
}

func joinGraphQLErrorMessages(errors []graphQLError) string {
	messages := make([]string, 0, len(errors))
	for _, graphQLError := range errors {
		if graphQLError.Message != "" {
			messages = append(messages, graphQLError.Message)
		}
	}

	return strings.Join(messages, "; ")
}

func messageContains(message string, needle string) bool {
	return strings.Contains(strings.ToLower(message), strings.ToLower(needle))
}
