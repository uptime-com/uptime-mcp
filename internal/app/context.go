package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"
)

// provideContext returns a context that is cancelled on SIGINT/SIGTERM.
func provideContext() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return ctx
}

// attachSignalHandler ensures graceful shutdown on signals.
func attachSignalHandler(lc fx.Lifecycle, ctx context.Context) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				<-ctx.Done()
			}()
			return nil
		},
	})
}
