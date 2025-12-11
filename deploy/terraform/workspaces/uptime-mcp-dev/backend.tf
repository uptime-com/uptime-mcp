terraform {
  cloud {
    organization = "uptime-com"
    workspaces {
      name = "uptime-mcp-dev"
    }
  }
}
