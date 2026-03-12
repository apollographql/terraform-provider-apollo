resource "apollo_subgraph_api_key" "products" {
  graph_id      = "inventory"
  variant       = "production"
  subgraph_name = "products"
  name          = "products-router"
}

import {
  to = apollo_subgraph_api_key.products
  id = "inventory:production:products:subgraph-key-123"
}
