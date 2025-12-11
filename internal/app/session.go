package app

import (
	"context"

	uptime "github.com/uptime-com/uptime-client-go"
)

type sessionKeyType struct{}

var sessionKey sessionKeyType

// Session holds per-request state including the authenticated Uptime client.
type Session struct {
	Client *uptime.Client
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
