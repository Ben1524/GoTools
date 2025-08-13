package context

import (
	"context"
	"os"
	"os/signal"
	"testing"
)

var siganChannel = make(chan os.Signal, 1)

func TestContextWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	select {
	case <-ctx.Done():
		t.Error("context should not be done yet")
	default:
	}

	cancel()

	select {
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Errorf("expected context to be canceled, got %v", ctx.Err())
		}
	default:
		t.Error("context should be done after cancel")
	}
}

func Exit() {
	signal.Notify()
}
