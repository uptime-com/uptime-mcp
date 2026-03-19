aws_region       = "eu-central-1"
eks_cluster_name = "upeks-prodeu"
ecr_repositories = [
  { name = "uptime-com/uptime-mcp/kustomize", mutability = "MUTABLE" },
  { name = "uptime-com/uptime-mcp/app/uptime-mcp", },
]
ref = {
  tag = "main"
}
