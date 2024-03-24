package timeutil

import (
	"context"
	"time"
)

type key int

var timeKey key = 0

// WithTime will attach the given time to the context.
func WithTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, timeKey, t)
}

func Now(ctx context.Context) time.Time {
	if t, ok := ctx.Value(timeKey).(time.Time); ok {
		return t
	}
	return time.Now()
}
