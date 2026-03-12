// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
)

const (
	operationCreateGraphAPIKey    = "CreateGraphAPIKey"
	operationGetGraphAPIKey       = "GetGraphAPIKey"
	operationUpdateGraphAPIKey    = "UpdateGraphAPIKey"
	operationDeleteGraphAPIKey    = "DeleteGraphAPIKey"
	operationResolveGraphOrgID    = "ResolveGraphOrganizationID"
	operationCreateSubgraphAPIKey = "CreateSubgraphAPIKey"
	operationGetSubgraphAPIKey    = "GetSubgraphAPIKey"
	operationUpdateSubgraphAPIKey = "UpdateSubgraphAPIKey"
	operationDeleteSubgraphAPIKey = "DeleteSubgraphAPIKey"
	subgraphAPIKeyExpectedKeyType = "SUBGRAPH"
	subgraphAPIKeyExpectedResType = "SUBGRAPH"
)

const (
	graphAPIKeySelection = `
      id
      keyName
      role
      createdAt
      lastUsed`

	graphAPIKeyCreateSelection = graphAPIKeySelection + `
      token`

	subgraphAPIKeySelection = `
      id
      keyName
      keyType
      createdAt
      lastUsed
      resources {
        resourceId
        resourceType
      }`

	subgraphAPIKeyCreateSelection = subgraphAPIKeySelection + `
      token`

	createGraphAPIKeyQuery = `
mutation CreateGraphAPIKey($graphId: ID!, $name: String!, $role: UserPermission!) {
  graph(id: $graphId) {
    newKey(keyName: $name, role: $role) {` + graphAPIKeyCreateSelection + `
    }
  }
}`

	getGraphAPIKeyQuery = `
query GetGraphAPIKey($graphId: ID!) {
  graph(id: $graphId) {
    id
    apiKeys {` + graphAPIKeySelection + `
    }
  }
}`

	updateGraphAPIKeyQuery = `
mutation UpdateGraphAPIKey($graphId: ID!, $id: ID!, $name: String!) {
  graph(id: $graphId) {
    renameKey(id: $id, newKeyName: $name) {` + graphAPIKeySelection + `
    }
  }
}`

	deleteGraphAPIKeyQuery = `
mutation DeleteGraphAPIKey($graphId: ID!, $id: ID!) {
  graph(id: $graphId) {
    removeKey(id: $id)
  }
}`

	resolveGraphOrganizationIDQuery = `
query ResolveGraphOrganizationID($graphId: ID!) {
  graph(id: $graphId) {
    id
    account {
      id
    }
  }
}`

	createSubgraphAPIKeyQuery = `
mutation CreateSubgraphAPIKey($organizationId: ID!, $name: String!, $resources: ApiKeyResourceInput!, $type: GraphOsKeyType!) {
  organization(id: $organizationId) {
    createKey(name: $name, resources: $resources, type: $type) {` + subgraphAPIKeyCreateSelection + `
    }
  }
}`

	getSubgraphAPIKeyQuery = `
query GetSubgraphAPIKey($organizationId: ID!, $keyId: ID!) {
  organization(id: $organizationId) {
    apiKey(keyId: $keyId) {` + subgraphAPIKeySelection + `
    }
  }
}`

	updateSubgraphAPIKeyQuery = `
mutation UpdateSubgraphAPIKey($organizationId: ID!, $keyId: ID!, $name: String!) {
  organization(id: $organizationId) {
    renameKey(keyId: $keyId, name: $name) {` + subgraphAPIKeySelection + `
    }
  }
}`

	deleteSubgraphAPIKeyQuery = `
mutation DeleteSubgraphAPIKey($organizationId: ID!, $keyId: ID!) {
  organization(id: $organizationId) {
    deleteKey(keyId: $keyId)
  }
}`
)

type GraphAPIKey struct {
	ID         string
	GraphID    string
	Name       string
	Role       string
	Token      string
	CreatedAt  string
	LastUsedAt string
}

type SubgraphAPIKey struct {
	ID             string
	OrganizationID string
	GraphID        string
	Variant        string
	SubgraphName   string
	Name           string
	Token          string
	CreatedAt      string
	LastUsedAt     string
}

type GraphAPIKeyInput struct {
	GraphID string
	ID      string
	Name    string
	Role    string
}

type SubgraphAPIKeyInput struct {
	GraphID      string
	Variant      string
	SubgraphName string
	ID           string
	Name         string
}

type graphAPIKeyPayload struct {
	ID        string `json:"id"`
	KeyName   string `json:"keyName"`
	Role      string `json:"role"`
	Token     string `json:"token"`
	CreatedAt string `json:"createdAt"`
	LastUsed  string `json:"lastUsed"`
}

type apiKeyResourcePayload struct {
	ResourceID   string `json:"resourceId"`
	ResourceType string `json:"resourceType"`
}

type subgraphAPIKeyPayload struct {
	ID        string                  `json:"id"`
	KeyName   string                  `json:"keyName"`
	KeyType   string                  `json:"keyType"`
	Token     string                  `json:"token"`
	CreatedAt string                  `json:"createdAt"`
	LastUsed  string                  `json:"lastUsed"`
	Resources []apiKeyResourcePayload `json:"resources"`
}

func (c *Client) CreateGraphAPIKey(ctx context.Context, input GraphAPIKeyInput) (*GraphAPIKey, error) {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.Name) == "" || NormalizeString(input.Role) == "" {
		return nil, inputError(operationCreateGraphAPIKey, "`graphID`, `name`, and `role` must be provided")
	}

	var response struct {
		Graph *struct {
			NewKey *graphAPIKeyPayload `json:"newKey"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationCreateGraphAPIKey, createGraphAPIKeyQuery, map[string]any{
		"graphId": input.GraphID,
		"name":    input.Name,
		"role":    input.Role,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil {
		return nil, classifyOperationError(operationCreateGraphAPIKey, 404, nil, fmt.Sprintf("graph %q was not found", input.GraphID), nil)
	}

	if response.Graph.NewKey == nil {
		return nil, malformedResponseError(operationCreateGraphAPIKey, fmt.Errorf("response did not include a graph API key"))
	}

	model := response.Graph.NewKey.toModel(input.GraphID)
	if model.Token == "" {
		return nil, malformedResponseError(operationCreateGraphAPIKey, fmt.Errorf("response did not include one-time key material"))
	}

	return &model, nil
}

func (c *Client) GetGraphAPIKey(ctx context.Context, graphID string, keyID string) (*GraphAPIKey, error) {
	if NormalizeString(graphID) == "" || NormalizeString(keyID) == "" {
		return nil, inputError(operationGetGraphAPIKey, "`graphID` and `keyID` must be provided")
	}

	var response struct {
		Graph *struct {
			APIKeys []graphAPIKeyPayload `json:"apiKeys"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationGetGraphAPIKey, getGraphAPIKeyQuery, map[string]any{
		"graphId": graphID,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil {
		return nil, classifyOperationError(operationGetGraphAPIKey, 404, nil, fmt.Sprintf("graph %q was not found", graphID), nil)
	}

	for _, payload := range response.Graph.APIKeys {
		if NormalizeString(payload.ID) == NormalizeString(keyID) {
			model := payload.toModel(graphID)
			return &model, nil
		}
	}

	return nil, classifyOperationError(operationGetGraphAPIKey, 404, nil, fmt.Sprintf("graph API key %q was not found on graph %q", keyID, graphID), nil)
}

func (c *Client) UpdateGraphAPIKey(ctx context.Context, input GraphAPIKeyInput) (*GraphAPIKey, error) {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.ID) == "" || NormalizeString(input.Name) == "" {
		return nil, inputError(operationUpdateGraphAPIKey, "`graphID`, `id`, and `name` must be provided")
	}

	var response struct {
		Graph *struct {
			RenameKey *graphAPIKeyPayload `json:"renameKey"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationUpdateGraphAPIKey, updateGraphAPIKeyQuery, map[string]any{
		"graphId": input.GraphID,
		"id":      input.ID,
		"name":    input.Name,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil || response.Graph.RenameKey == nil {
		return nil, classifyOperationError(operationUpdateGraphAPIKey, 404, nil, fmt.Sprintf("graph API key %q was not found on graph %q", input.ID, input.GraphID), nil)
	}

	model := response.Graph.RenameKey.toModel(input.GraphID)
	return &model, nil
}

func (c *Client) DeleteGraphAPIKey(ctx context.Context, graphID string, keyID string) error {
	if NormalizeString(graphID) == "" || NormalizeString(keyID) == "" {
		return inputError(operationDeleteGraphAPIKey, "`graphID` and `keyID` must be provided")
	}

	var response struct {
		Graph *struct {
			RemoveKey any `json:"removeKey"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationDeleteGraphAPIKey, deleteGraphAPIKeyQuery, map[string]any{
		"graphId": graphID,
		"id":      keyID,
	}, &response); err != nil {
		return err
	}

	if response.Graph == nil {
		return classifyOperationError(operationDeleteGraphAPIKey, 404, nil, fmt.Sprintf("graph %q was not found", graphID), nil)
	}

	return nil
}

func (c *Client) ResolveGraphOrganizationID(ctx context.Context, graphID string) (string, error) {
	if NormalizeString(graphID) == "" {
		return "", inputError(operationResolveGraphOrgID, "`graphID` must be provided")
	}

	var response struct {
		Graph *struct {
			Account *struct {
				ID string `json:"id"`
			} `json:"account"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationResolveGraphOrgID, resolveGraphOrganizationIDQuery, map[string]any{
		"graphId": graphID,
	}, &response); err != nil {
		return "", err
	}

	if response.Graph == nil {
		return "", classifyOperationError(operationResolveGraphOrgID, 404, nil, fmt.Sprintf("graph %q was not found", graphID), nil)
	}

	if response.Graph.Account == nil || NormalizeString(response.Graph.Account.ID) == "" {
		return "", malformedResponseError(operationResolveGraphOrgID, fmt.Errorf("response did not include the owning organization"))
	}

	return NormalizeString(response.Graph.Account.ID), nil
}

func (c *Client) CreateSubgraphAPIKey(ctx context.Context, input SubgraphAPIKeyInput) (*SubgraphAPIKey, error) {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.Variant) == "" || NormalizeString(input.SubgraphName) == "" || NormalizeString(input.Name) == "" {
		return nil, inputError(operationCreateSubgraphAPIKey, "`graphID`, `variant`, `subgraphName`, and `name` must be provided")
	}

	organizationID, err := c.ResolveGraphOrganizationID(ctx, input.GraphID)
	if err != nil {
		return nil, err
	}

	var response struct {
		Organization *struct {
			CreateKey *subgraphAPIKeyPayload `json:"createKey"`
		} `json:"organization"`
	}

	if err := c.doGraphQL(ctx, operationCreateSubgraphAPIKey, createSubgraphAPIKeyQuery, map[string]any{
		"organizationId": organizationID,
		"name":           input.Name,
		"type":           subgraphAPIKeyExpectedKeyType,
		"resources": map[string]any{
			"subgraphs": []map[string]string{
				{
					"graphId":      input.GraphID,
					"variantName":  input.Variant,
					"subgraphName": input.SubgraphName,
				},
			},
		},
	}, &response); err != nil {
		return nil, err
	}

	if response.Organization == nil {
		return nil, classifyOperationError(operationCreateSubgraphAPIKey, 404, nil, fmt.Sprintf("organization %q was not found", organizationID), nil)
	}

	key, err := subgraphAPIKeyFromPayload(operationCreateSubgraphAPIKey, organizationID, response.Organization.CreateKey)
	if err != nil {
		return nil, err
	}

	if err := key.matchesTuple(operationCreateSubgraphAPIKey, input.GraphID, input.Variant, input.SubgraphName); err != nil {
		return nil, err
	}

	if key.Token == "" {
		return nil, malformedResponseError(operationCreateSubgraphAPIKey, fmt.Errorf("response did not include one-time key material"))
	}

	return key, nil
}

func (c *Client) GetSubgraphAPIKey(ctx context.Context, input SubgraphAPIKeyInput) (*SubgraphAPIKey, error) {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.Variant) == "" || NormalizeString(input.SubgraphName) == "" || NormalizeString(input.ID) == "" {
		return nil, inputError(operationGetSubgraphAPIKey, "`graphID`, `variant`, `subgraphName`, and `id` must be provided")
	}

	organizationID, err := c.ResolveGraphOrganizationID(ctx, input.GraphID)
	if err != nil {
		return nil, err
	}

	var response struct {
		Organization *struct {
			APIKey *subgraphAPIKeyPayload `json:"apiKey"`
		} `json:"organization"`
	}

	if err := c.doGraphQL(ctx, operationGetSubgraphAPIKey, getSubgraphAPIKeyQuery, map[string]any{
		"organizationId": organizationID,
		"keyId":          input.ID,
	}, &response); err != nil {
		return nil, err
	}

	if response.Organization == nil || response.Organization.APIKey == nil {
		return nil, classifyOperationError(operationGetSubgraphAPIKey, 404, nil, fmt.Sprintf("subgraph API key %q was not found", input.ID), nil)
	}

	key, err := subgraphAPIKeyFromPayload(operationGetSubgraphAPIKey, organizationID, response.Organization.APIKey)
	if err != nil {
		return nil, err
	}

	if err := key.matchesTuple(operationGetSubgraphAPIKey, input.GraphID, input.Variant, input.SubgraphName); err != nil {
		return nil, err
	}

	return key, nil
}

func (c *Client) UpdateSubgraphAPIKey(ctx context.Context, input SubgraphAPIKeyInput) (*SubgraphAPIKey, error) {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.Variant) == "" || NormalizeString(input.SubgraphName) == "" || NormalizeString(input.ID) == "" || NormalizeString(input.Name) == "" {
		return nil, inputError(operationUpdateSubgraphAPIKey, "`graphID`, `variant`, `subgraphName`, `id`, and `name` must be provided")
	}

	organizationID, err := c.ResolveGraphOrganizationID(ctx, input.GraphID)
	if err != nil {
		return nil, err
	}

	var response struct {
		Organization *struct {
			RenameKey *subgraphAPIKeyPayload `json:"renameKey"`
		} `json:"organization"`
	}

	if err := c.doGraphQL(ctx, operationUpdateSubgraphAPIKey, updateSubgraphAPIKeyQuery, map[string]any{
		"organizationId": organizationID,
		"keyId":          input.ID,
		"name":           input.Name,
	}, &response); err != nil {
		return nil, err
	}

	if response.Organization == nil || response.Organization.RenameKey == nil {
		return nil, classifyOperationError(operationUpdateSubgraphAPIKey, 404, nil, fmt.Sprintf("subgraph API key %q was not found", input.ID), nil)
	}

	key, err := subgraphAPIKeyFromPayload(operationUpdateSubgraphAPIKey, organizationID, response.Organization.RenameKey)
	if err != nil {
		return nil, err
	}

	if err := key.matchesTuple(operationUpdateSubgraphAPIKey, input.GraphID, input.Variant, input.SubgraphName); err != nil {
		return nil, err
	}

	return key, nil
}

func (c *Client) DeleteSubgraphAPIKey(ctx context.Context, input SubgraphAPIKeyInput) error {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.ID) == "" {
		return inputError(operationDeleteSubgraphAPIKey, "`graphID` and `id` must be provided")
	}

	organizationID, err := c.ResolveGraphOrganizationID(ctx, input.GraphID)
	if err != nil {
		return err
	}

	var response struct {
		Organization *struct {
			DeleteKey string `json:"deleteKey"`
		} `json:"organization"`
	}

	if err := c.doGraphQL(ctx, operationDeleteSubgraphAPIKey, deleteSubgraphAPIKeyQuery, map[string]any{
		"organizationId": organizationID,
		"keyId":          input.ID,
	}, &response); err != nil {
		return err
	}

	if response.Organization == nil {
		return classifyOperationError(operationDeleteSubgraphAPIKey, 404, nil, fmt.Sprintf("organization %q was not found", organizationID), nil)
	}

	if NormalizeString(response.Organization.DeleteKey) == "" {
		return malformedResponseError(operationDeleteSubgraphAPIKey, fmt.Errorf("response did not include a deleted key identifier"))
	}

	if NormalizeString(input.ID) != NormalizeString(response.Organization.DeleteKey) {
		return malformedResponseError(operationDeleteSubgraphAPIKey, fmt.Errorf("response deleted key %q instead of %q", response.Organization.DeleteKey, input.ID))
	}

	return nil
}

func (p graphAPIKeyPayload) toModel(graphID string) GraphAPIKey {
	return GraphAPIKey{
		ID:         NormalizeString(p.ID),
		GraphID:    NormalizeString(graphID),
		Name:       NormalizeString(p.KeyName),
		Role:       NormalizeString(p.Role),
		Token:      NormalizeString(p.Token),
		CreatedAt:  NormalizeString(p.CreatedAt),
		LastUsedAt: NormalizeString(p.LastUsed),
	}
}

func subgraphAPIKeyFromPayload(operation string, organizationID string, payload *subgraphAPIKeyPayload) (*SubgraphAPIKey, error) {
	if payload == nil {
		return nil, malformedResponseError(operation, fmt.Errorf("response did not include a subgraph API key"))
	}

	if strings.ToUpper(NormalizeString(payload.KeyType)) != subgraphAPIKeyExpectedKeyType {
		return nil, malformedResponseError(operation, fmt.Errorf("response returned key type %q instead of %q", payload.KeyType, subgraphAPIKeyExpectedKeyType))
	}

	if len(payload.Resources) == 0 {
		return nil, malformedResponseError(operation, fmt.Errorf("response did not include a subgraph target"))
	}

	if len(payload.Resources) > 1 {
		return nil, capabilityDeferredError(operation, fmt.Sprintf("`apollo_subgraph_api_key` supports exactly one subgraph target, but GraphOS returned %d targets for key %q.", len(payload.Resources), payload.ID))
	}

	resource := payload.Resources[0]
	if strings.ToUpper(NormalizeString(resource.ResourceType)) != subgraphAPIKeyExpectedResType {
		return nil, malformedResponseError(operation, fmt.Errorf("response returned resource type %q instead of %q", resource.ResourceType, subgraphAPIKeyExpectedResType))
	}

	graphID, variant, subgraphName, err := parseSubgraphAPIKeyResourceID(resource.ResourceID)
	if err != nil {
		return nil, malformedResponseError(operation, fmt.Errorf("unexpected subgraph resource identifier %q: %w", resource.ResourceID, err))
	}

	return &SubgraphAPIKey{
		ID:             NormalizeString(payload.ID),
		OrganizationID: NormalizeString(organizationID),
		GraphID:        graphID,
		Variant:        variant,
		SubgraphName:   subgraphName,
		Name:           NormalizeString(payload.KeyName),
		Token:          NormalizeString(payload.Token),
		CreatedAt:      NormalizeString(payload.CreatedAt),
		LastUsedAt:     NormalizeString(payload.LastUsed),
	}, nil
}

func parseSubgraphAPIKeyResourceID(resourceID string) (string, string, string, error) {
	parts, err := ParseFlexibleImportID(resourceID, 3, 4)
	if err != nil {
		return "", "", "", err
	}

	switch len(parts) {
	case 3:
		return parts[0], parts[1], parts[2], nil
	case 4:
		return parts[1], parts[2], parts[3], nil
	default:
		return "", "", "", fmt.Errorf("expected 3 or 4 colon-delimited parts, got %d", len(parts))
	}
}

func (k *SubgraphAPIKey) matchesTuple(operation string, graphID string, variant string, subgraphName string) error {
	if k == nil {
		return malformedResponseError(operation, fmt.Errorf("response did not include a subgraph API key"))
	}

	expected := JoinImportID(graphID, variant, subgraphName)
	actual := JoinImportID(k.GraphID, k.Variant, k.SubgraphName)
	if actual != expected {
		return inputError(operation, fmt.Sprintf("subgraph API key %q targets %s, not %s", k.ID, actual, expected))
	}

	return nil
}
