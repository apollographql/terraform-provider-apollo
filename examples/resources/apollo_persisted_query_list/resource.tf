resource "apollo_persisted_query_list" "mobile" {
  graph_id    = "inventory"
  name        = "mobile-clients"
  description = "Persisted queries used by the mobile applications."
}
