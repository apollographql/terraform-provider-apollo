resource "apollo_session_policy" "org" {
  organization_id            = "apollo-org"
  max_session_length_minutes = 480
}
