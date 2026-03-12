// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net/url"
	"strings"
)

type StringNormalizer func(string) string

func NormalizeString(value string) string {
	return strings.TrimSpace(value)
}

func SemanticallyEqual(left string, right string, normalizer StringNormalizer) bool {
	if normalizer == nil {
		normalizer = NormalizeString
	}

	return normalizer(left) == normalizer(right)
}

func NormalizeURLString(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""

	if (parsed.Scheme == "http" && strings.HasSuffix(parsed.Host, ":80")) || (parsed.Scheme == "https" && strings.HasSuffix(parsed.Host, ":443")) {
		hostParts := strings.Split(parsed.Host, ":")
		parsed.Host = hostParts[0]
	}

	if parsed.Path == "/" {
		parsed.Path = ""
	}

	return parsed.String()
}
