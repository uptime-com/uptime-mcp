terraform {
  cloud {
    organization = "uptime-com"
    workspaces {
      name = "uptime-mcp-prodeu"
    }
  }
}
