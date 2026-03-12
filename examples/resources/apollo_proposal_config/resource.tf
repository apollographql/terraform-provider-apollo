resource "apollo_proposal_config" "production" {
  graph_id          = "inventory"
  variant           = "production"
  enabled           = true
  require_approvals = true
}
