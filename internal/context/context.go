package context

import (
	"context"

	"go.uber.org/atomic"
)

type ctxKey int

const (
	ctxRequestId ctxKey = iota
)

var (
	requestID atomic.Uint64
)

func WithRequestId(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxRequestId, requestID.Inc())
}

func GetRequestId(ctx context.Context) uint64 {
	reqId, ok := ctx.Value(ctxRequestId).(uint64)
	if !ok {
		return 0
	}
	return reqId
}
