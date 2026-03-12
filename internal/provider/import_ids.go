// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
)

func ParseImportID(id string, expectedParts int) ([]string, error) {
	return ParseFlexibleImportID(id, expectedParts, expectedParts)
}

func ParseFlexibleImportID(id string, minParts int, maxParts int) ([]string, error) {
	if minParts <= 0 || maxParts < minParts {
		return nil, fmt.Errorf("invalid import ID parser configuration")
	}

	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil, fmt.Errorf("import ID cannot be empty")
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) < minParts || len(parts) > maxParts {
		if minParts == maxParts {
			return nil, fmt.Errorf("expected import ID with %d part(s), got %d", minParts, len(parts))
		}

		return nil, fmt.Errorf("expected import ID with %d to %d part(s), got %d", minParts, maxParts, len(parts))
	}

	for index, part := range parts {
		parts[index] = strings.TrimSpace(part)
		if parts[index] == "" {
			return nil, fmt.Errorf("import ID part %d cannot be empty", index+1)
		}
	}

	return parts, nil
}

func JoinImportID(parts ...string) string {
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return strings.Join(normalized, ":")
}
