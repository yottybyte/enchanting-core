package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/yottybyte/enchanting-core/internal/config"
)

var mapConfigLevels map[string]slog.Level = map[string]slog.Level{
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
	"debug": slog.LevelDebug,
}

func New(cfg *config.Logger) (*slog.Logger, error) {
	l, ok := mapConfigLevels[cfg.Level]
	if !ok {
		return nil, fmt.Errorf("unknown logger level: %s", cfg.Level)
	}

	var handlerOptions = slog.HandlerOptions{
		Level: l,
	}

	switch cfg.Type {
	case "text":
		return slog.New(slog.NewTextHandler(os.Stdout, &handlerOptions)), nil
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, &handlerOptions)), nil
	default:
		return nil, fmt.Errorf("unknown logger type: %s", cfg.Type)
	}
}
