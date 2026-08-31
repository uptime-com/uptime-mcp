package app

import (
	"context"

	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

type sessionKeyType struct{}

var sessionKey sessionKeyType

// AuthScheme names one of the two Authorization schemes api/v1 accepts. They
// are not interchangeable: an OAuth2 access token presented as Token, or an
// account API key presented as Bearer, is refused as NOT_AUTHENTICATED.
type AuthScheme string

const (
	// SchemeBearer authenticates an OAuth2 access token.
	SchemeBearer AuthScheme = "Bearer"
	// SchemeToken authenticates a static account API key.
	SchemeToken AuthScheme = "Token"
)

// Other returns the scheme a credential belongs to if it does not belong to s.
func (s AuthScheme) Other() AuthScheme {
	if s == SchemeBearer {
		return SchemeToken
	}
	return SchemeBearer
}

// Session holds per-session state including the authenticated Uptime client.
// Client is created once per session by middleware and cached.
//
// Scheme is where the credential came from rather than anything read off
// Token: the two credentials are both opaque strings, and which one a given
// string is cannot be recovered from it.
type Session struct {
	Token  string
	Scheme AuthScheme
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
