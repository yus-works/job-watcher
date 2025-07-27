// logging/ctxlog.go
package logging

import (
	"context"
	"log/slog"
	"os"
)

type key struct{}

var Level = new(slog.LevelVar) // defaults to INFO (0)

func Init() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: Level,
	})
	slog.SetDefault(slog.New(h))
}

func Into(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, key{}, l)
}

func From(ctx context.Context) *slog.Logger {
	if v := ctx.Value(key{}); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}
