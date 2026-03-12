resource "apollo_graph_api_key" "ci" {
  graph_id = "inventory"
  name     = "inventory-ci"
  role     = "GRAPH_ADMIN"
}

import {
  to = apollo_graph_api_key.ci
  id = "inventory:graph-key-123"
}
