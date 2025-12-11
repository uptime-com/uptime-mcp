module "awsecr" {
  source     = "git::ssh://git@github.com/uptime-com/uptf.git//tfmods/awsecr?ref=v0.4"
  for_each   = { for repo in var.ecr_repositories : repo.name => repo }
  name       = each.value.name
  mutability = each.value.mutability
}

resource "kubernetes_namespace_v1" "this" {
  metadata {
    name = "uptime-mcp"
  }
  wait_for_default_service_account = true
}

module "fluxcd" {
  source            = "git::ssh://git@github.com/uptime-com/uptfmods.git//fluxcd/repo?ref=v0.71"
  namespace         = one(kubernetes_namespace_v1.this.metadata.*.name)
  name              = "uptime-mcp"
  github_repository = "uptime-mcp"
  ref = {
    tag = "main"
  }
  reconcile_interval = "5m"
}
