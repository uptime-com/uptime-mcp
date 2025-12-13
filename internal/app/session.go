package app

import (
	"context"

	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

type sessionKeyType struct{}

var sessionKey sessionKeyType

// Session holds per-session state including the authenticated Uptime client.
// Client is created once per session by middleware and cached.
type Session struct {
	APIKey string
	Client upapi.API
}

// ContextWithSession returns a context with session attached.
func ContextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

// SessionFromContext retrieves session from context.
func SessionFromContext(ctx context.Context) *Session {
	session, _ := ctx.Value(sessionKey).(*Session)
	return session
}
