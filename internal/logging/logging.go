// logging/ctxlog.go
package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type key struct{}

var Level = new(slog.LevelVar) // defaults to INFO (0)

func Init() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     Level,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key != slog.SourceKey {
				return a
			}
			switch v := a.Value.Any().(type) {
			case slog.Source:
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", filepath.Base(v.File), v.Line))
			case *slog.Source:
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", filepath.Base(v.File), v.Line))
			case string:
				// v already "path/file.go:123" -> strip path
				if i := strings.LastIndexByte(v, '/'); i >= 0 {
					v = v[i+1:]
				}
				a.Value = slog.StringValue(v)
			}
			return a
		},
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
