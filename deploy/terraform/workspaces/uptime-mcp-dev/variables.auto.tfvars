aws_region       = "us-east-2"
eks_cluster_name = "upeks-dev"
ecr_repositories = [
  { name = "uptime-com/uptime-mcp/kustomize", mutability = "MUTABLE" },
  { name = "uptime-com/uptime-mcp/app/uptime-mcp", },
]
ref = {
  tag = "main"
}
