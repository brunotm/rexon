package rexon

import (
	"context"
)

// wrapCtxSend wraps the sending to a channel with a context
func wrapCtxSend(ctx context.Context, result Result, resultCh chan<- Result) (ok bool) {
	select {
	case <-ctx.Done():
		return false
	case resultCh <- result:
		return true
	}
}
