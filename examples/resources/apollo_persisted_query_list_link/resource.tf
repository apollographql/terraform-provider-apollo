resource "apollo_persisted_query_list" "mobile" {
  graph_id = "inventory"
  name     = "mobile-clients"
}

resource "apollo_persisted_query_list_link" "current" {
  persisted_query_list_id = apollo_persisted_query_list.mobile.id
  graph_id                = apollo_persisted_query_list.mobile.graph_id
  variant                 = "current"
}
