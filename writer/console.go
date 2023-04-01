package writer

import (
	"fmt"
	"os"
)

type ConsoleWriter struct {}

func (w *ConsoleWriter) Write(b []byte) (int, error) {
    fmt.Fprintln(os.Stdout, string(b))
    return 0, nil
}
