package server

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func loggingMiddleware(w io.Writer) mcp.Middleware {
	logger := log.New(w, "", 0)
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			name := method
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				name = ctr.Params.Name
			}
			t := time.Now()
			defer logger.Printf("%s %d", name, time.Since(t).Milliseconds())
			return next(ctx, method, req)
		}
	}
}
