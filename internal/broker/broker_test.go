package broker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/officialasishkumar/pulsemesh/internal/observability"
)

func TestPublishAndProcess(t *testing.T) {
	b := New(4, 2, observability.NewMetrics())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	processed := make(chan Message, 1)
	b.Start(ctx, 1, HandlerFunc(func(_ context.Context, msg Message) error {
		processed <- msg
		return nil
	}))

	if err := b.Publish(observability.WithTraceID(ctx, "trace-1"), Message{Topic: "orders", Payload: "created"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case msg := <-processed:
		if msg.TraceID != "trace-1" {
			t.Fatalf("trace id was not propagated: %q", msg.TraceID)
		}
	case <-time.After(time.Second):
		t.Fatal("message was not processed")
	}
}

func TestPublishReturnsQueueFull(t *testing.T) {
	b := New(1, 1, observability.NewMetrics())
	ctx := context.Background()
	if err := b.Publish(ctx, Message{Topic: "a", Payload: "1"}); err != nil {
		t.Fatalf("first publish failed: %v", err)
	}
	if err := b.Publish(ctx, Message{Topic: "b", Payload: "2"}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}

func TestFailedMessageMovesToDeadLetterQueue(t *testing.T) {
	b := New(1, 1, observability.NewMetrics())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var attempts atomic.Int32
	b.Start(ctx, 1, HandlerFunc(func(context.Context, Message) error {
		attempts.Add(1)
		return errors.New("downstream unavailable")
	}))
	if err := b.Publish(ctx, Message{Topic: "payments", Payload: "capture"}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		if got := b.DeadLetters(); len(got) == 1 {
			if attempts.Load() != 3 {
				t.Fatalf("expected 3 attempts, got %d", attempts.Load())
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("message did not reach dead letter queue")
		default:
			time.Sleep(25 * time.Millisecond)
		}
	}
}
