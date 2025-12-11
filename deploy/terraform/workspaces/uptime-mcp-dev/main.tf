module "awsecr" {
  source     = "git::ssh://git@github.com/uptime-com/uptf.git//tfmods/awsecr?ref=v0.4"
  for_each   = { for repo in var.ecr_repositories : repo.name => repo }
  name       = each.value.name
  mutability = each.value.mutability
}
