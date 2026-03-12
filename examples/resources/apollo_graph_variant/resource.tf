resource "apollo_graph_variant" "production" {
  graph_id = "inventory"
  variant  = "production"
}

import {
  to = apollo_graph_variant.production
  id = "inventory:production"
}
