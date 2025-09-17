package study

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func Context() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	logger.Info("Application started", slog.String("env", "production"), slog.Int("port", 8080))

	select {
	case <-ctx.Done():
		fmt.Println("Context done")
	case <-time.After(time.Second * 3):
		fmt.Println("Timeout done")
	}
}
