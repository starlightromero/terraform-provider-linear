resource "linear_team_membership" "example" {
  team_id = linear_team.example.id
  user_id = "00000000-0000-0000-0000-000000000000"
  owner   = false
}
