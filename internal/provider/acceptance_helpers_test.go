// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

type testServerMode string

const (
	testServerModeNormal           testServerMode = ""
	testServerModeUnauthenticated  testServerMode = "unauthenticated"
	testServerModePermissionDenied testServerMode = "permission_denied"
)

func requireHermeticAcceptance(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("set TF_ACC=1 to run hermetic acceptance tests")
	}
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"apollo": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func variableString(variables map[string]any, key string) string {
	value, ok := variables[key]
	if !ok || value == nil {
		return ""
	}

	return strings.TrimSpace(fmt.Sprint(value))
}

func writeJSONResponse(t *testing.T, w http.ResponseWriter, payload map[string]any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("failed to encode JSON response: %s", err)
	}
}
