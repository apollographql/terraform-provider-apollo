// Copyright (c) Apollo Graph, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

type CapabilityState string

const (
	CapabilitySupported  CapabilityState = "supported"
	CapabilityDeferred   CapabilityState = "deferred"
	CapabilityOutOfScope CapabilityState = "out_of_scope"
)

type ResourceCapability struct {
	Create CapabilityState
	Read   CapabilityState
	Update CapabilityState
	Delete CapabilityState
	Import CapabilityState
	Notes  string
}

var resourceCapabilities = map[string]ResourceCapability{
	"graph_variant": {
		Create: CapabilityDeferred,
		Read:   CapabilitySupported,
		Update: CapabilityDeferred,
		Delete: CapabilitySupported,
		Import: CapabilitySupported,
		Notes:  "New variants are created by publishing the first subgraph, not by a standalone variant-create mutation. The current public Platform API schema does expose GraphVariantMutation.delete for brownfield cleanup.",
	},
	"subgraph": {
		Create: CapabilityDeferred,
		Read:   CapabilitySupported,
		Update: CapabilityDeferred,
		Delete: CapabilitySupported,
		Import: CapabilitySupported,
		Notes:  "Brownfield subgraph support is import/read/delete only. Creation still requires schema publish input, so Terraform does not bootstrap or mutate registrations in this slice.",
	},
	"persisted_query_list": {
		Create: CapabilitySupported,
		Read:   CapabilitySupported,
		Update: CapabilitySupported,
		Delete: CapabilitySupported,
		Import: CapabilitySupported,
		Notes:  "The current public Platform API schema confirms graph-scoped persisted query list create, read-by-id, updateMetadata, and delete operations. The current API accepts description writes but does not expose description for readback.",
	},
	"persisted_query_list_link": {
		Create: CapabilitySupported,
		Read:   CapabilitySupported,
		Update: CapabilitySupported,
		Delete: CapabilitySupported,
		Import: CapabilitySupported,
		Notes:  "The current public Platform API schema exposes variant-scoped linkPersistedQueryList and unlinkPersistedQueryList mutations under GraphVariantMutation.",
	},
	"graph_api_key": {
		Create: CapabilitySupported,
		Read:   CapabilitySupported,
		Update: CapabilitySupported,
		Delete: CapabilitySupported,
		Import: CapabilitySupported,
		Notes:  "The current public Platform API schema exposes graph(id).newKey, graph(id).apiKeys, graph(id).renameKey, and graph(id).removeKey. The API key value is only available on create, and role changes require replacement because GraphOS does not allow changing graph API key roles after creation.",
	},
	"subgraph_api_key": {
		Create: CapabilitySupported,
		Read:   CapabilitySupported,
		Update: CapabilitySupported,
		Delete: CapabilitySupported,
		Import: CapabilitySupported,
		Notes:  "The current public Platform API schema exposes organization(id).createKey, apiKey, renameKey, and deleteKey for subgraph API keys. This provider slice supports exactly one subgraph target tuple per resource and preserves the one-time key value only for Terraform-created resources.",
	},
}

func ResourceCapabilityForType(typeName string) (ResourceCapability, bool) {
	capability, ok := resourceCapabilities[typeName]
	return capability, ok
}
