package server

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func loggingMiddleware(w io.Writer) mcp.Middleware {
	logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey || a.Key == slog.MessageKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			start := time.Now()

			attrs := []slog.Attr{
				slog.String("method", method),
			}

			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				attrs = append(attrs, slog.String("tool", ctr.Params.Name))
			}

			if req != nil {
				if extra := req.GetExtra(); extra != nil && extra.Header != nil {
					if ip := extra.Header.Get("X-Forwarded-For"); ip != "" {
						attrs = append(attrs, slog.String("client_ip", ip))
					} else if ip := extra.Header.Get("X-Real-IP"); ip != "" {
						attrs = append(attrs, slog.String("client_ip", ip))
					}
					if ua := extra.Header.Get("User-Agent"); ua != "" {
						attrs = append(attrs, slog.String("user_agent", ua))
					}
				}
			}

			result, err := next(ctx, method, req)

			latency := time.Since(start)
			attrs = append(attrs,
				slog.String("latency", latency.String()),
				slog.Int64("latency_ms", latency.Milliseconds()),
			)
			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
			}

			logger.LogAttrs(ctx, slog.LevelInfo, "request", attrs...)
			return result, err
		}
	}
}
