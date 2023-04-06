package writer

import (
	"context"
	"fmt"
	"os"
)

type ConsoleWriter struct {}

func (w *ConsoleWriter) Write(ctx context.Context, b []byte) error {
    fmt.Fprintln(os.Stdout, string(b))
    return nil
}
