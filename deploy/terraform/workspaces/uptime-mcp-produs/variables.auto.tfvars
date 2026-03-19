aws_region       = "us-east-2"
eks_cluster_name = "upeks-produs"
ecr_repositories = [
  { name = "uptime-com/uptime-mcp/kustomize", mutability = "MUTABLE" },
  { name = "uptime-com/uptime-mcp/app/uptime-mcp", },
]
ref = {
  tag = "main"
}
