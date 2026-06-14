package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextWithSession(t *testing.T) {
	t.Run("attaches session to context", func(t *testing.T) {
		session := &Session{Token: "test-token"}
		ctx := ContextWithSession(context.Background(), session)

		require.NotNil(t, ctx)

		retrieved := SessionFromContext(ctx)
		require.NotNil(t, retrieved)
		assert.Equal(t, "test-token", retrieved.Token)
		assert.Same(t, session, retrieved)
	})
}

func TestSessionFromContext(t *testing.T) {
	t.Run("returns session when present", func(t *testing.T) {
		session := &Session{Token: "my-token"}
		ctx := ContextWithSession(context.Background(), session)

		retrieved := SessionFromContext(ctx)

		require.NotNil(t, retrieved)
		assert.Equal(t, "my-token", retrieved.Token)
	})

	t.Run("returns nil when no session", func(t *testing.T) {
		ctx := context.Background()

		retrieved := SessionFromContext(ctx)

		assert.Nil(t, retrieved)
	})

	t.Run("returns nil for wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), sessionKey, "not-a-session")

		retrieved := SessionFromContext(ctx)

		assert.Nil(t, retrieved)
	})
}
