package serr

import (
	"context"
	"log/slog"
	"os"
)

func ExampleLogError() {
	err := New("sample error", String("attr1", "abcd"))
	LogError(context.Background(), slog.New(slog.NewJSONHandler(os.Stdout, nil)), err)
}
