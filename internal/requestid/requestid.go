package requestid

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

func Set(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxRequestId, requestID.Inc())
}

func Get(ctx context.Context) uint64 {
	if ctx == nil {
		return 0
	}
	reqId, ok := ctx.Value(ctxRequestId).(uint64)
	if !ok {
		return 0
	}
	return reqId
}
