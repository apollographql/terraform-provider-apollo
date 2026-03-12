data "apollo_subgraph" "products" {
  graph_id = "inventory"
  variant  = "production"
  name     = "products"
}
