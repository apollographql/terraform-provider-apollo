// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func AddAPIErrorDiagnostic(diags *diag.Diagnostics, action string, err error, requiredScope string) {
	var apiErr *APIError
	if !AsAPIError(err, &apiErr) {
		diags.AddError(action, err.Error())
		return
	}

	summary := action
	detail := apiErr.Error()

	switch apiErr.Kind {
	case ErrorKindUnauthenticated:
		summary = fmt.Sprintf("%s: authentication failed", action)
		detail = "The GraphOS Platform API rejected the configured credential. Check `api_key` or `APOLLO_KEY` and verify that the token is still valid."
	case ErrorKindPermissionDenied:
		summary = fmt.Sprintf("%s: permission denied", action)
		detail = "The configured GraphOS credential does not have enough authority for this operation."
		if requiredScope != "" {
			detail = fmt.Sprintf("%s Expected credential scope: %s.", detail, requiredScope)
		}
	case ErrorKindNotFound:
		summary = fmt.Sprintf("%s: remote object not found", action)
	case ErrorKindRateLimited:
		summary = fmt.Sprintf("%s: GraphOS rate limit exceeded", action)
	case ErrorKindMalformedResponse:
		summary = fmt.Sprintf("%s: malformed GraphOS response", action)
	case ErrorKindCapabilityDeferred:
		summary = fmt.Sprintf("%s: capability deferred", action)
	}

	diags.AddError(summary, detail)
}
