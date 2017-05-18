package rexon

import (
	"context"

	log "github.com/Sirupsen/logrus"
)

var (
	rexParseSize        = RexCompile(`([-+]?[0-9]*\.?[0-9]+)\s*(\w+)?`)
	rexRemoveEmptyLines = RexCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	emptyByte           = []byte("")
)

// RemoveEmptyLines from the given []byte
func RemoveEmptyLines(b []byte) []byte {
	return rexRemoveEmptyLines.ReplaceAll(b, emptyByte)
}

// wrapCtxSend wraps the sending to a channel with a context
func wrapCtxSend(ctx context.Context, document []byte, rexChan chan<- []byte) bool {
	select {
	case <-ctx.Done():
		log.WithField("document", document).Warn("rexon: timed out sending on channel")
		return false
	case rexChan <- document:
		return true
	}
}
