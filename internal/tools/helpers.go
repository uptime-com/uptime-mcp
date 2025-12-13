package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// textResult creates a successful text response.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
