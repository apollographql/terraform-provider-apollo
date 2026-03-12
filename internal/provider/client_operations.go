// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
)

const (
	operationCreateGraphVariant            = "CreateGraphVariant"
	operationGetGraphMetadata              = "GetGraphMetadata"
	operationGetVariantMetadata            = "GetVariantMetadata"
	operationDeleteGraphVariant            = "DeleteGraphVariant"
	operationGetSubgraphMetadata           = "GetSubgraphMetadata"
	operationRemoveSubgraph                = "RemoveSubgraph"
	operationPublishSubgraphSchema         = "PublishSubgraphSchema"
	operationCreatePersistedQueryList      = "CreatePersistedQueryList"
	operationGetPersistedQueryList         = "GetPersistedQueryList"
	operationGetVariantPersistedQueryList  = "GetVariantPersistedQueryList"
	operationListPersistedQueryLists       = "ListPersistedQueryLists"
	operationUpdatePersistedQueryList      = "UpdatePersistedQueryList"
	operationDeletePersistedQueryList      = "DeletePersistedQueryList"
	operationLinkPersistedQueryListVariant = "LinkPersistedQueryListToVariant"
	operationUnlinkPersistedQueryList      = "UnlinkPersistedQueryListFromVariant"
)

// Queries in this file are confirmed either by Apollo's public docs or by the
// current public GraphOS Platform API schema.
const (
	persistedQueryListSelection = `
      id
      name
      linkedVariants {
        name
        graph {
          id
        }
      }`

	getGraphMetadataQuery = `
query GetGraphMetadata($graphId: ID!) {
  graph(id: $graphId) {
    id
    name
    title
    variants {
      name
    }
  }
}`

	getVariantMetadataQuery = `
query GetVariantMetadata($graphId: ID!, $variantName: String!) {
  graph(id: $graphId) {
    id
    variant(name: $variantName) {
      id
      name
      subgraphs {
        name
        url
        revision
      }
    }
  }
}`

	deleteGraphVariantQuery = `
mutation DeleteGraphVariant($graphId: ID!, $variantName: String!) {
  graph(id: $graphId) {
    variant(name: $variantName) {
      delete {
        deleted
      }
    }
  }
}`

	// The public docs confirm schema publication and first-variant bootstrap via publishSubgraph.
	publishSubgraphQuery = `
mutation PublishSubgraphSchema($graphId: ID!, $variant: String!, $subgraph: String!, $url: String!, $schema: String!, $revision: String!) {
  graph(id: $graphId) {
    publishSubgraph(graphVariant: $variant, name: $subgraph, url: $url, revision: $revision, activePartialSchema: { sdl: $schema }) {
      __typename
    }
  }
}`

	removeSubgraphQuery = `
mutation RemoveSubgraph($graphId: ID!, $variant: String!, $subgraph: String!) {
  graph(id: $graphId) {
    removeImplementingServiceAndTriggerComposition(graphVariant: $variant, name: $subgraph, dryRun: false) {
      updatedGateway
      errors {
        code
        message
      }
    }
  }
}`

	getPersistedQueryListQuery = `
query GetPersistedQueryList($graphId: ID!, $id: ID!) {
  graph(id: $graphId) {
    persistedQueryList(id: $id) {` + persistedQueryListSelection + `
    }
  }
}`

	getVariantPersistedQueryListQuery = `
query GetVariantPersistedQueryList($graphId: ID!, $variantName: String!) {
  graph(id: $graphId) {
    variant(name: $variantName) {
      persistedQueryList {` + persistedQueryListSelection + `
      }
    }
  }
}`

	createPersistedQueryListQuery = `
mutation CreatePersistedQueryList($graphId: ID!, $name: String!, $description: String) {
  graph(id: $graphId) {
    createPersistedQueryList(name: $name, description: $description) {
      __typename
      ... on PermissionError {
        message
      }
      ... on CreatePersistedQueryListResult {
        persistedQueryList {` + persistedQueryListSelection + `
        }
      }
    }
  }
}`

	updatePersistedQueryListQuery = `
mutation UpdatePersistedQueryList($graphId: ID!, $id: ID!, $name: String!, $description: String) {
  graph(id: $graphId) {
    persistedQueryList(id: $id) {
      updateMetadata(name: $name, description: $description) {
        __typename
        ... on PermissionError {
          message
        }
        ... on UpdatePersistedQueryListMetadataResult {
          persistedQueryList {` + persistedQueryListSelection + `
          }
        }
      }
    }
  }
}`

	deletePersistedQueryListQuery = `
mutation DeletePersistedQueryList($graphId: ID!, $id: ID!) {
  graph(id: $graphId) {
    persistedQueryList(id: $id) {
      delete {
        __typename
        ... on PermissionError {
          message
        }
        ... on CannotDeleteLinkedPersistedQueryListError {
          message
        }
        ... on DeletePersistedQueryListResult {
          graph {
            id
          }
        }
      }
    }
  }
}`

	linkPersistedQueryListToVariantQuery = `
mutation LinkPersistedQueryListToVariant($graphId: ID!, $variantName: String!, $persistedQueryListId: ID!) {
  graph(id: $graphId) {
    variant(name: $variantName) {
      linkPersistedQueryList(persistedQueryListId: $persistedQueryListId) {
        __typename
        ... on PermissionError {
          message
        }
        ... on ListNotFoundError {
          message
          listId
        }
        ... on VariantAlreadyLinkedError {
          message
        }
      }
    }
  }
}`

	unlinkPersistedQueryListFromVariantQuery = `
mutation UnlinkPersistedQueryListFromVariant($graphId: ID!, $variantName: String!) {
  graph(id: $graphId) {
    variant(name: $variantName) {
      unlinkPersistedQueryList {
        __typename
        ... on PermissionError {
          message
        }
        ... on VariantAlreadyUnlinkedError {
          message
        }
      }
    }
  }
}`
)

type GraphMetadata struct {
	ID           string
	Name         string
	Title        string
	VariantNames []string
}

type GraphVariantRef struct {
	GraphID string
	Variant string
}

type VariantMetadata struct {
	ID        string
	GraphID   string
	Name      string
	Subgraphs []SubgraphMetadata
}

type SubgraphMetadata struct {
	GraphID    string
	Variant    string
	Name       string
	RoutingURL string
	Revision   string
}

type PublishSubgraphInput struct {
	GraphID    string
	Variant    string
	Subgraph   string
	RoutingURL string
	SchemaSDL  string
	Revision   string
}

type PersistedQueryList struct {
	ID             string
	Name           string
	LinkedVariants []GraphVariantRef
}

type PersistedQueryListInput struct {
	GraphID     string
	ID          string
	Name        string
	Description string
}

type PersistedQueryListLinkInput struct {
	PersistedQueryListID string
	GraphID              string
	Variant              string
}

func (c *Client) GetGraphMetadata(ctx context.Context, graphID string) (*GraphMetadata, error) {
	if NormalizeString(graphID) == "" {
		return nil, inputError(operationGetGraphMetadata, "`graphID` must be provided")
	}

	var response struct {
		Graph *struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Title    string `json:"title"`
			Variants []struct {
				Name string `json:"name"`
			} `json:"variants"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationGetGraphMetadata, getGraphMetadataQuery, map[string]any{
		"graphId": graphID,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil {
		return nil, classifyOperationError(operationGetGraphMetadata, 404, nil, fmt.Sprintf("graph %q was not found", graphID), nil)
	}

	result := &GraphMetadata{
		ID:    response.Graph.ID,
		Name:  response.Graph.Name,
		Title: response.Graph.Title,
	}
	for _, variant := range response.Graph.Variants {
		result.VariantNames = append(result.VariantNames, variant.Name)
	}

	return result, nil
}

func (c *Client) GetVariantMetadata(ctx context.Context, graphID string, variantName string) (*VariantMetadata, error) {
	if NormalizeString(graphID) == "" || NormalizeString(variantName) == "" {
		return nil, inputError(operationGetVariantMetadata, "`graphID` and `variantName` must be provided")
	}

	var response struct {
		Graph *struct {
			ID      string `json:"id"`
			Variant *struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Subgraphs []struct {
					Name     string `json:"name"`
					URL      string `json:"url"`
					Revision string `json:"revision"`
				} `json:"subgraphs"`
			} `json:"variant"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationGetVariantMetadata, getVariantMetadataQuery, map[string]any{
		"graphId":     graphID,
		"variantName": variantName,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil || response.Graph.Variant == nil {
		return nil, classifyOperationError(operationGetVariantMetadata, 404, nil, fmt.Sprintf("variant %q on graph %q was not found", variantName, graphID), nil)
	}

	result := &VariantMetadata{
		ID:      response.Graph.Variant.ID,
		GraphID: response.Graph.ID,
		Name:    response.Graph.Variant.Name,
	}

	for _, subgraph := range response.Graph.Variant.Subgraphs {
		result.Subgraphs = append(result.Subgraphs, SubgraphMetadata{
			GraphID:    response.Graph.ID,
			Variant:    response.Graph.Variant.Name,
			Name:       subgraph.Name,
			RoutingURL: subgraph.URL,
			Revision:   subgraph.Revision,
		})
	}

	return result, nil
}

func (c *Client) GetSubgraphMetadata(ctx context.Context, graphID string, variantName string, subgraphName string) (*SubgraphMetadata, error) {
	if NormalizeString(subgraphName) == "" {
		return nil, inputError(operationGetSubgraphMetadata, "`subgraphName` must be provided")
	}

	variant, err := c.GetVariantMetadata(ctx, graphID, variantName)
	if err != nil {
		return nil, err
	}

	for _, subgraph := range variant.Subgraphs {
		if subgraph.Name == subgraphName {
			result := subgraph
			return &result, nil
		}
	}

	return nil, classifyOperationError(operationGetSubgraphMetadata, 404, nil, fmt.Sprintf("subgraph %q was not found on %s", subgraphName, JoinImportID(graphID, variantName)), nil)
}

func (c *Client) DeleteGraphVariant(ctx context.Context, graphID string, variantName string) error {
	if NormalizeString(graphID) == "" || NormalizeString(variantName) == "" {
		return inputError(operationDeleteGraphVariant, "`graphID` and `variantName` must be provided")
	}

	var response struct {
		Graph *struct {
			Variant *struct {
				Delete *struct {
					Deleted bool `json:"deleted"`
				} `json:"delete"`
			} `json:"variant"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationDeleteGraphVariant, deleteGraphVariantQuery, map[string]any{
		"graphId":     graphID,
		"variantName": variantName,
	}, &response); err != nil {
		return err
	}

	if response.Graph == nil || response.Graph.Variant == nil {
		return classifyOperationError(operationDeleteGraphVariant, 404, nil, fmt.Sprintf("variant %q on graph %q was not found", variantName, graphID), nil)
	}

	if response.Graph.Variant.Delete == nil {
		return malformedResponseError(operationDeleteGraphVariant, fmt.Errorf("response did not include a delete result"))
	}

	if !response.Graph.Variant.Delete.Deleted {
		return malformedResponseError(operationDeleteGraphVariant, fmt.Errorf("response did not confirm deleted=true"))
	}

	return nil
}

func (c *Client) RemoveSubgraph(ctx context.Context, graphID string, variantName string, subgraphName string) error {
	if NormalizeString(graphID) == "" || NormalizeString(variantName) == "" || NormalizeString(subgraphName) == "" {
		return inputError(operationRemoveSubgraph, "`graphID`, `variantName`, and `subgraphName` must be provided")
	}

	var response struct {
		Graph *struct {
			RemoveSubgraph *struct {
				UpdatedGateway bool `json:"updatedGateway"`
				Errors         []struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"errors"`
			} `json:"removeImplementingServiceAndTriggerComposition"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationRemoveSubgraph, removeSubgraphQuery, map[string]any{
		"graphId":  graphID,
		"variant":  variantName,
		"subgraph": subgraphName,
	}, &response); err != nil {
		if (IsErrorKind(err, ErrorKindUnknown) || IsErrorKind(err, ErrorKindTransient)) && subgraphWasRemoved(ctx, c, graphID, variantName, subgraphName) {
			return nil
		}

		return err
	}

	if response.Graph == nil {
		return classifyOperationError(operationRemoveSubgraph, 404, nil, fmt.Sprintf("graph %q was not found", graphID), nil)
	}

	if response.Graph.RemoveSubgraph == nil {
		return malformedResponseError(operationRemoveSubgraph, fmt.Errorf("response did not include a subgraph removal result"))
	}

	if len(response.Graph.RemoveSubgraph.Errors) > 0 {
		return classifyOperationError(
			operationRemoveSubgraph,
			400,
			nil,
			fmt.Sprintf(
				"subgraph removal for %q on %s produced composition errors: %s",
				subgraphName,
				JoinImportID(graphID, variantName),
				joinSubgraphCompositionErrorMessages(response.Graph.RemoveSubgraph.Errors),
			),
			nil,
		)
	}

	if !response.Graph.RemoveSubgraph.UpdatedGateway {
		return classifyOperationError(
			operationRemoveSubgraph,
			400,
			nil,
			fmt.Sprintf("subgraph removal for %q on %s did not update the gateway", subgraphName, JoinImportID(graphID, variantName)),
			nil,
		)
	}

	return nil
}

func subgraphWasRemoved(ctx context.Context, client *Client, graphID string, variantName string, subgraphName string) bool {
	if client == nil {
		return false
	}

	_, err := client.GetSubgraphMetadata(ctx, graphID, variantName, subgraphName)
	return IsErrorKind(err, ErrorKindNotFound)
}

func joinSubgraphCompositionErrorMessages(errors []struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}) string {
	messages := make([]string, 0, len(errors))
	for _, compositionError := range errors {
		message := NormalizeString(compositionError.Message)
		if message == "" {
			continue
		}

		if code := NormalizeString(compositionError.Code); code != "" {
			message = fmt.Sprintf("%s (%s)", message, code)
		}

		messages = append(messages, message)
	}

	if len(messages) == 0 {
		return "unknown composition error"
	}

	return strings.Join(messages, "; ")
}

func (c *Client) PublishSubgraphSchema(ctx context.Context, input PublishSubgraphInput) error {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.Variant) == "" || NormalizeString(input.Subgraph) == "" || NormalizeString(input.RoutingURL) == "" || NormalizeString(input.SchemaSDL) == "" || NormalizeString(input.Revision) == "" {
		return inputError(operationPublishSubgraphSchema, "`graphID`, `variant`, `subgraph`, `routingURL`, `schemaSDL`, and `revision` must all be provided")
	}

	return c.doGraphQL(ctx, operationPublishSubgraphSchema, publishSubgraphQuery, map[string]any{
		"graphId":  input.GraphID,
		"variant":  input.Variant,
		"subgraph": input.Subgraph,
		"url":      input.RoutingURL,
		"schema":   input.SchemaSDL,
		"revision": input.Revision,
	}, &struct {
		Graph map[string]any `json:"graph"`
	}{})
}

func (c *Client) CreatePersistedQueryList(ctx context.Context, input PersistedQueryListInput) (*PersistedQueryList, error) {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.Name) == "" {
		return nil, inputError(operationCreatePersistedQueryList, "`graphID` and `name` must be provided")
	}

	var response struct {
		Graph *struct {
			CreatePersistedQueryList *persistedQueryListResultPayload `json:"createPersistedQueryList"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationCreatePersistedQueryList, createPersistedQueryListQuery, map[string]any{
		"graphId":     input.GraphID,
		"name":        input.Name,
		"description": nullIfEmpty(input.Description),
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil {
		return nil, classifyOperationError(operationCreatePersistedQueryList, 404, nil, fmt.Sprintf("graph %q was not found", input.GraphID), nil)
	}

	return persistedQueryListResultFromPayload(operationCreatePersistedQueryList, response.Graph.CreatePersistedQueryList)
}

func (c *Client) GetPersistedQueryList(ctx context.Context, graphID string, id string) (*PersistedQueryList, error) {
	if NormalizeString(graphID) == "" || NormalizeString(id) == "" {
		return nil, inputError(operationGetPersistedQueryList, "`graphID` and `id` must be provided")
	}

	var response struct {
		Graph *struct {
			PersistedQueryList *persistedQueryListPayload `json:"persistedQueryList"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationGetPersistedQueryList, getPersistedQueryListQuery, map[string]any{
		"graphId": graphID,
		"id":      id,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil || response.Graph.PersistedQueryList == nil {
		return nil, classifyOperationError(operationGetPersistedQueryList, 404, nil, fmt.Sprintf("persisted query list %q was not found on graph %q", id, graphID), nil)
	}

	model := response.Graph.PersistedQueryList.toModel()
	return &model, nil
}

func (c *Client) GetVariantPersistedQueryList(ctx context.Context, graphID string, variantName string) (*PersistedQueryList, error) {
	if NormalizeString(graphID) == "" || NormalizeString(variantName) == "" {
		return nil, inputError(operationGetVariantPersistedQueryList, "`graphID` and `variantName` must be provided")
	}

	var response struct {
		Graph *struct {
			Variant *struct {
				PersistedQueryList *persistedQueryListPayload `json:"persistedQueryList"`
			} `json:"variant"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationGetVariantPersistedQueryList, getVariantPersistedQueryListQuery, map[string]any{
		"graphId":     graphID,
		"variantName": variantName,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil || response.Graph.Variant == nil {
		return nil, classifyOperationError(operationGetVariantPersistedQueryList, 404, nil, fmt.Sprintf("variant %q on graph %q was not found", variantName, graphID), nil)
	}

	if response.Graph.Variant.PersistedQueryList == nil {
		return nil, nil
	}

	model := response.Graph.Variant.PersistedQueryList.toModel()
	return &model, nil
}

func (c *Client) ListPersistedQueryLists(_ context.Context, graphID string) ([]PersistedQueryList, error) {
	if NormalizeString(graphID) == "" {
		return nil, inputError(operationListPersistedQueryLists, "`graphID` must be provided")
	}

	return nil, capabilityDeferredError(operationListPersistedQueryLists, "The current public GraphOS Platform API schema does not expose `graph(id).persistedQueryLists`, so listing persisted query lists is not currently available.")
}

func (c *Client) UpdatePersistedQueryList(ctx context.Context, input PersistedQueryListInput) (*PersistedQueryList, error) {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.ID) == "" || NormalizeString(input.Name) == "" {
		return nil, inputError(operationUpdatePersistedQueryList, "`graphID`, `id`, and `name` must be provided")
	}

	var response struct {
		Graph *struct {
			PersistedQueryList *struct {
				UpdateMetadata *persistedQueryListResultPayload `json:"updateMetadata"`
			} `json:"persistedQueryList"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationUpdatePersistedQueryList, updatePersistedQueryListQuery, map[string]any{
		"graphId":     input.GraphID,
		"id":          input.ID,
		"name":        input.Name,
		"description": nullIfEmpty(input.Description),
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil || response.Graph.PersistedQueryList == nil {
		return nil, classifyOperationError(operationUpdatePersistedQueryList, 404, nil, fmt.Sprintf("persisted query list %q was not found on graph %q", input.ID, input.GraphID), nil)
	}

	return persistedQueryListResultFromPayload(operationUpdatePersistedQueryList, response.Graph.PersistedQueryList.UpdateMetadata)
}

func (c *Client) DeletePersistedQueryList(ctx context.Context, graphID string, id string) error {
	if NormalizeString(graphID) == "" || NormalizeString(id) == "" {
		return inputError(operationDeletePersistedQueryList, "`graphID` and `id` must be provided")
	}

	var response struct {
		Graph *struct {
			PersistedQueryList *struct {
				Delete *persistedQueryListMutationPayload `json:"delete"`
			} `json:"persistedQueryList"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationDeletePersistedQueryList, deletePersistedQueryListQuery, map[string]any{
		"graphId": graphID,
		"id":      id,
	}, &response); err != nil {
		return err
	}

	if response.Graph == nil || response.Graph.PersistedQueryList == nil {
		return classifyOperationError(operationDeletePersistedQueryList, 404, nil, fmt.Sprintf("persisted query list %q was not found on graph %q", id, graphID), nil)
	}

	if response.Graph.PersistedQueryList.Delete == nil {
		return malformedResponseError(operationDeletePersistedQueryList, fmt.Errorf("response did not include a delete result"))
	}

	return response.Graph.PersistedQueryList.Delete.errorForType(operationDeletePersistedQueryList, deletePersistedQueryListSuccessTypeName)
}

func (c *Client) LinkPersistedQueryListToVariant(ctx context.Context, input PersistedQueryListLinkInput) (*PersistedQueryList, error) {
	if NormalizeString(input.PersistedQueryListID) == "" || NormalizeString(input.GraphID) == "" || NormalizeString(input.Variant) == "" {
		return nil, inputError(operationLinkPersistedQueryListVariant, "`persistedQueryListID`, `graphID`, and `variant` must be provided")
	}

	current, err := c.GetVariantPersistedQueryList(ctx, input.GraphID, input.Variant)
	if err != nil && !IsErrorKind(err, ErrorKindNotFound) {
		return nil, err
	}

	if current != nil {
		if current.ID == input.PersistedQueryListID && persistedQueryListHasVariantLink(current, input.GraphID, input.Variant) {
			return current, nil
		}

		return nil, inputError(operationLinkPersistedQueryListVariant, fmt.Sprintf("variant %q on graph %q is already linked to persisted query list %q", input.Variant, input.GraphID, current.ID))
	}

	var response struct {
		Graph *struct {
			Variant *struct {
				LinkPersistedQueryList *persistedQueryListLinkPayload `json:"linkPersistedQueryList"`
			} `json:"variant"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationLinkPersistedQueryListVariant, linkPersistedQueryListToVariantQuery, map[string]any{
		"graphId":              input.GraphID,
		"variantName":          input.Variant,
		"persistedQueryListId": input.PersistedQueryListID,
	}, &response); err != nil {
		return nil, err
	}

	if response.Graph == nil || response.Graph.Variant == nil {
		return nil, classifyOperationError(operationLinkPersistedQueryListVariant, 404, nil, fmt.Sprintf("variant %q on graph %q was not found", input.Variant, input.GraphID), nil)
	}

	if payload := response.Graph.Variant.LinkPersistedQueryList; payload != nil {
		if err := payload.errorForType(operationLinkPersistedQueryListVariant); err != nil {
			return nil, err
		}
	}

	linked, err := c.GetVariantPersistedQueryList(ctx, input.GraphID, input.Variant)
	if err != nil {
		return nil, err
	}

	if linked == nil || linked.ID != input.PersistedQueryListID || !persistedQueryListHasVariantLink(linked, input.GraphID, input.Variant) {
		return nil, malformedResponseError(operationLinkPersistedQueryListVariant, fmt.Errorf("link mutation completed without the requested persisted query list association"))
	}

	return linked, nil
}

func (c *Client) UnlinkPersistedQueryListFromVariant(ctx context.Context, input PersistedQueryListLinkInput) error {
	if NormalizeString(input.GraphID) == "" || NormalizeString(input.Variant) == "" {
		return inputError(operationUnlinkPersistedQueryList, "`graphID` and `variant` must be provided")
	}

	current, err := c.GetVariantPersistedQueryList(ctx, input.GraphID, input.Variant)
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			return nil
		}

		return err
	}

	if current == nil {
		return nil
	}

	if NormalizeString(input.PersistedQueryListID) != "" && current.ID != input.PersistedQueryListID {
		return nil
	}

	var response struct {
		Graph *struct {
			Variant *struct {
				UnlinkPersistedQueryList *persistedQueryListMutationPayload `json:"unlinkPersistedQueryList"`
			} `json:"variant"`
		} `json:"graph"`
	}

	if err := c.doGraphQL(ctx, operationUnlinkPersistedQueryList, unlinkPersistedQueryListFromVariantQuery, map[string]any{
		"graphId":     input.GraphID,
		"variantName": input.Variant,
	}, &response); err != nil {
		return err
	}

	if response.Graph == nil || response.Graph.Variant == nil {
		return nil
	}

	if payload := response.Graph.Variant.UnlinkPersistedQueryList; payload != nil {
		if err := classifyPersistedQueryListResultType(operationUnlinkPersistedQueryList, payload.Typename, payload.Message); err != nil {
			if messageContains(err.Error(), "already unlinked") {
				return nil
			}

			return err
		}
	}

	remaining, err := c.GetVariantPersistedQueryList(ctx, input.GraphID, input.Variant)
	if err != nil {
		if IsErrorKind(err, ErrorKindNotFound) {
			return nil
		}

		return err
	}

	if remaining != nil && NormalizeString(input.PersistedQueryListID) != "" && remaining.ID == input.PersistedQueryListID {
		return malformedResponseError(operationUnlinkPersistedQueryList, fmt.Errorf("unlink mutation completed but the persisted query list is still associated with the variant"))
	}

	return nil
}

type persistedQueryListPayload struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LinkedVariants []struct {
		Name  string `json:"name"`
		Graph struct {
			ID string `json:"id"`
		} `json:"graph"`
	} `json:"linkedVariants"`
}

func (p persistedQueryListPayload) toModel() PersistedQueryList {
	result := PersistedQueryList{
		ID:   p.ID,
		Name: p.Name,
	}

	for _, linkedVariant := range p.LinkedVariants {
		result.LinkedVariants = append(result.LinkedVariants, GraphVariantRef{
			GraphID: linkedVariant.Graph.ID,
			Variant: linkedVariant.Name,
		})
	}

	return result
}

type persistedQueryListMutationPayload struct {
	Typename string `json:"__typename"`
	Message  string `json:"message"`
}

type persistedQueryListResultPayload struct {
	Typename           string                     `json:"__typename"`
	Message            string                     `json:"message"`
	PersistedQueryList *persistedQueryListPayload `json:"persistedQueryList"`
}

type persistedQueryListLinkPayload struct {
	Typename string `json:"__typename"`
	Message  string `json:"message"`
	ListID   string `json:"listId"`
}

func (p persistedQueryListMutationPayload) errorForType(operation string, successTypename string) error {
	return classifyPersistedQueryListResultType(operation, p.Typename, p.Message, successTypename)
}

func (p persistedQueryListLinkPayload) errorForType(operation string) error {
	return classifyPersistedQueryListResultType(operation, p.Typename, p.Message)
}

const (
	createPersistedQueryListSuccessTypeName = "CreatePersistedQueryListResult"
	updatePersistedQueryListSuccessTypeName = "UpdatePersistedQueryListMetadataResult"
	deletePersistedQueryListSuccessTypeName = "DeletePersistedQueryListResult"
)

func classifyPersistedQueryListResultType(operation string, typename string, message string, successTypenames ...string) error {
	normalizedTypename := NormalizeString(typename)
	switch normalizedTypename {
	case "":
		return malformedResponseError(operation, fmt.Errorf("response did not include __typename"))
	case "PermissionError":
		return classifyOperationError(operation, 403, nil, message, nil)
	case "ValidationError", "InvalidRefFormat", "InvalidTarget", "VariantAlreadyLinkedError", "VariantAlreadyUnlinkedError", "CannotDeleteLinkedPersistedQueryListError":
		return classifyOperationError(operation, 400, nil, message, nil)
	case "NotFoundError", "ListNotFoundError":
		return classifyOperationError(operation, 404, nil, message, nil)
	}

	for _, successTypename := range successTypenames {
		if normalizedTypename == successTypename {
			return nil
		}
	}

	if len(successTypenames) == 0 {
		return nil
	}

	return malformedResponseError(operation, fmt.Errorf("unexpected result type %q", normalizedTypename))
}

func persistedQueryListResultFromPayload(operation string, payload *persistedQueryListResultPayload) (*PersistedQueryList, error) {
	if payload == nil {
		return nil, malformedResponseError(operation, fmt.Errorf("response did not include a result payload"))
	}

	successTypename := createPersistedQueryListSuccessTypeName
	if operation == operationUpdatePersistedQueryList {
		successTypename = updatePersistedQueryListSuccessTypeName
	}

	if err := classifyPersistedQueryListResultType(operation, payload.Typename, payload.Message, successTypename); err != nil {
		return nil, err
	}

	if payload.PersistedQueryList == nil {
		return nil, malformedResponseError(operation, fmt.Errorf("response did not include a persisted query list"))
	}

	model := payload.PersistedQueryList.toModel()
	return &model, nil
}

func persistedQueryListHasVariantLink(pql *PersistedQueryList, graphID string, variant string) bool {
	if pql == nil {
		return false
	}

	for _, linkedVariant := range pql.LinkedVariants {
		if NormalizeString(linkedVariant.GraphID) == NormalizeString(graphID) && NormalizeString(linkedVariant.Variant) == NormalizeString(variant) {
			return true
		}
	}

	return false
}

func nullIfEmpty(value string) any {
	if NormalizeString(value) == "" {
		return nil
	}

	return value
}
