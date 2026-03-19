module "awsecr" {
  source     = "git::ssh://git@github.com/uptime-com/uptf.git//tfmods/awsecr?ref=v0.4"
  for_each   = { for repo in var.ecr_repositories : repo.name => repo }
  name       = each.value.name
  mutability = each.value.mutability
}

module "fluxcd" {
  source             = "git::ssh://git@github.com/uptime-com/uptfmods.git//fluxcd/repo?ref=v0.71"
  github_repository  = "uptime-com/uptime-mcp"
  namespace          = "uptime-mcp"
  name               = "uptime-mcp"
  ref                = var.ref
  reconcile_interval = "5m"
}
