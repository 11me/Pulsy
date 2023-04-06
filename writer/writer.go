package writer

import "context"

type Writer interface{
    Write(context.Context, []byte) error
}

