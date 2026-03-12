resource "apollo_subgraph" "products" {
  graph_id = "inventory"
  variant  = "production"
  name     = "products"
}

import {
  to = apollo_subgraph.products
  id = "inventory:production:products"
}
