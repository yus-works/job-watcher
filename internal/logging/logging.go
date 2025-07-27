package logging

import (
	"log/slog"
	"os"
)

var Level = new(slog.LevelVar) // default INFO

func New() *slog.Logger {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: Level})
	return slog.New(h)
}
