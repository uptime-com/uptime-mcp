package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"
)

// ProvideContext returns a context that is cancelled on SIGINT/SIGTERM.
func ProvideContext() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return ctx
}

// AttachSignalHandler ensures graceful shutdown on signals.
func AttachSignalHandler(lc fx.Lifecycle, ctx context.Context) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				<-ctx.Done()
			}()
			return nil
		},
	})
}
