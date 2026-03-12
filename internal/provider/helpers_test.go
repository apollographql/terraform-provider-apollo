// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestParseImportID(t *testing.T) {
	t.Parallel()

	parts, err := ParseImportID("graph:variant", 2)
	if err != nil {
		t.Fatalf("expected import ID to parse, got error: %s", err)
	}

	if got, want := strings.Join(parts, "|"), "graph|variant"; got != want {
		t.Fatalf("unexpected import parts: got %q want %q", got, want)
	}

	if _, err := ParseImportID("graph::variant", 3); err == nil {
		t.Fatal("expected empty import ID segment to fail")
	}
}

func TestParseFlexibleImportID(t *testing.T) {
	t.Parallel()

	parts, err := ParseFlexibleImportID("pql:graph:variant", 2, 3)
	if err != nil {
		t.Fatalf("expected flexible import ID to parse, got error: %s", err)
	}

	if got, want := len(parts), 3; got != want {
		t.Fatalf("unexpected number of parts: got %d want %d", got, want)
	}
}

func TestNormalizeURLString(t *testing.T) {
	t.Parallel()

	left := " HTTPS://EXAMPLE.COM:443/ "
	right := "https://example.com"

	if !SemanticallyEqual(left, right, NormalizeURLString) {
		t.Fatalf("expected URLs to normalize to the same value: %q vs %q", left, right)
	}
}

func TestAddAPIErrorDiagnostic(t *testing.T) {
	t.Parallel()

	var diagnostics diag.Diagnostics
	AddAPIErrorDiagnostic(&diagnostics, "create persisted query list", &APIError{
		Kind:      ErrorKindPermissionDenied,
		Operation: operationCreatePersistedQueryList,
		Message:   "forbidden",
	}, "org-admin")

	if !diagnosticsContainSummary(diagnostics, "permission denied") {
		t.Fatalf("expected permission denied diagnostic, got %#v", diagnostics)
	}

	if !strings.Contains(diagnostics[0].Detail(), "org-admin") {
		t.Fatalf("expected required scope in detail, got %q", diagnostics[0].Detail())
	}
}
