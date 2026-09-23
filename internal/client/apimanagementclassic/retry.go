package apimanagementclassic

import (
	"context"
	"math/rand"
	"time"
)

// pollUntilVisible repeatedly calls check until it returns true, a
// non-nil error, or ctx's own deadline/cancellation fires — whichever
// comes first. It exists for exactly one purpose: SAP's own documentation
// for API Provider Create/Update/Delete states that "changes may not be
// immediately reflected in subsequent GET API provider requests... typically
// available within approximately 20 seconds," caused by response caching in
// front of the destination service. A bare time.Sleep(20 * time.Second)
// would either under-wait (SAP never promises an upper bound, only a
// typical figure) or waste time when the backend is already consistent;
// polling with capped exponential backoff and jitter, bounded by ctx,
// adapts to both cases and never blocks longer than the caller allows.
func pollUntilVisible(ctx context.Context, check func(context.Context) (bool, error)) error {
	const (
		initialDelay = 500 * time.Millisecond
		maxDelay     = 5 * time.Second
	)

	delay := initialDelay
	for {
		visible, err := check(ctx)
		if err != nil {
			return err
		}
		if visible {
			return nil
		}

		jitter := time.Duration(rand.Int63n(int64(delay) / 2)) //nolint:gosec // G404: jitter timing, not security-sensitive
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay + jitter):
		}

		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}
