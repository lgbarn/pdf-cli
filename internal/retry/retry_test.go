package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDoSuccess(t *testing.T) {
	t.Parallel()
	calls := 0
	err := Do(context.Background(), Options{MaxAttempts: 3, BaseDelay: time.Millisecond}, func(_ context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDoRetryThenSuccess(t *testing.T) {
	t.Parallel()
	calls := 0
	err := Do(context.Background(), Options{MaxAttempts: 3, BaseDelay: time.Millisecond}, func(_ context.Context) error {
		calls++
		if calls < 2 {
			return errors.New("transient error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDoExhaustion(t *testing.T) {
	t.Parallel()
	calls := 0
	sentinel := errors.New("always fails")
	err := Do(context.Background(), Options{MaxAttempts: 3, BaseDelay: time.Millisecond}, func(_ context.Context) error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDoPermanentError(t *testing.T) {
	t.Parallel()
	calls := 0
	inner := errors.New("not found")
	err := Do(context.Background(), Options{MaxAttempts: 5, BaseDelay: time.Millisecond}, func(_ context.Context) error {
		calls++
		return Permanent(inner)
	})
	if !errors.Is(err, inner) {
		t.Fatalf("expected inner error via Unwrap, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call (no retry on permanent), got %d", calls)
	}
}

func TestDoContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := Do(ctx, Options{MaxAttempts: 10, BaseDelay: 50 * time.Millisecond}, func(_ context.Context) error {
		calls++
		if calls == 2 {
			cancel()
		}
		return errors.New("keep going")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDoBackoffTiming(t *testing.T) {
	t.Parallel()
	baseDelay := 10 * time.Millisecond
	calls := 0
	starts := make([]time.Time, 0, 4)

	_ = Do(context.Background(), Options{MaxAttempts: 4, BaseDelay: baseDelay, MaxDelay: time.Second}, func(_ context.Context) error {
		starts = append(starts, time.Now())
		calls++
		return errors.New("fail")
	})

	if calls != 4 {
		t.Fatalf("expected 4 calls, got %d", calls)
	}

	// Expected delays: 10ms, 20ms, 40ms (between attempts 0-1, 1-2, 2-3)
	expectedDelays := []time.Duration{baseDelay, 2 * baseDelay, 4 * baseDelay}
	tolerance := 8 * time.Millisecond

	for i := 0; i < len(expectedDelays); i++ {
		actual := starts[i+1].Sub(starts[i])
		expected := expectedDelays[i]
		if actual < expected-tolerance || actual > expected+3*tolerance {
			t.Errorf("delay %d: expected ~%v, got %v", i, expected, actual)
		}
	}
}

func TestIsPermanent(t *testing.T) {
	t.Parallel()

	t.Run("direct", func(t *testing.T) {
		err := Permanent(errors.New("bad"))
		if !IsPermanent(err) {
			t.Fatal("expected IsPermanent to return true for direct PermanentError")
		}
	})

	t.Run("wrapped", func(t *testing.T) {
		inner := Permanent(errors.New("bad"))
		wrapped := fmt.Errorf("outer: %w", inner)
		if !IsPermanent(wrapped) {
			t.Fatal("expected IsPermanent to return true for wrapped PermanentError")
		}
	})

	t.Run("not_permanent", func(t *testing.T) {
		err := errors.New("regular error")
		if IsPermanent(err) {
			t.Fatal("expected IsPermanent to return false for regular error")
		}
	})

	t.Run("nil", func(t *testing.T) {
		if IsPermanent(nil) {
			t.Fatal("expected IsPermanent to return false for nil")
		}
	})
}
