package logger

import (
	"log/slog"
	"os"
)

// Setup initializes the global slog logger with a simple text handler
func Setup() {
	// Create a text handler writing to stdout
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		AddSource: true,
	})
	
	// Set the default slog logger
	slog.SetDefault(slog.New(handler))
}
