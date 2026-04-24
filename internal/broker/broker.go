package broker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/officialasishkumar/senior-systems-lab/internal/observability"
)

var (
	ErrQueueFull = errors.New("queue is full")
	ErrClosed    = errors.New("queue is closed")
)

type Message struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Payload   string            `json:"payload"`
	TraceID   string            `json:"trace_id"`
	Attempts  int               `json:"attempts"`
	Headers   map[string]string `json:"headers,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type Handler interface {
	Handle(context.Context, Message) error
}

type HandlerFunc func(context.Context, Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}

type Broker struct {
	queue   chan Message
	dlq     chan Message
	metrics *observability.Metrics
	closed  atomic.Bool
	ids     atomic.Uint64
	once    sync.Once
}

func New(capacity int, deadLetterCapacity int, metrics *observability.Metrics) *Broker {
	return &Broker{
		queue:   make(chan Message, capacity),
		dlq:     make(chan Message, deadLetterCapacity),
		metrics: metrics,
	}
}

func (b *Broker) Publish(ctx context.Context, msg Message) error {
	if b.closed.Load() {
		return ErrClosed
	}
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg-%d", b.ids.Add(1))
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}
	if msg.TraceID == "" {
		msg.TraceID = observability.TraceID(ctx)
	}
	select {
	case b.queue <- msg:
		b.metrics.IncQueuePublished()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrQueueFull
	}
}

func (b *Broker) Start(ctx context.Context, workers int, handler Handler) {
	if workers < 1 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go b.worker(ctx, handler)
	}
}

func (b *Broker) worker(ctx context.Context, handler Handler) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-b.queue:
			if !ok {
				return
			}
			b.handle(ctx, handler, msg)
		}
	}
}

func (b *Broker) handle(ctx context.Context, handler Handler, msg Message) {
	traceCtx := observability.WithTraceID(ctx, msg.TraceID)
	for attempt := 1; attempt <= 3; attempt++ {
		msg.Attempts = attempt
		if err := handler.Handle(traceCtx, msg); err == nil {
			b.metrics.IncQueueProcessed()
			return
		}
		b.metrics.IncQueueFailed()
		time.Sleep(time.Duration(attempt) * 25 * time.Millisecond)
	}
	select {
	case b.dlq <- msg:
		b.metrics.IncQueueDeadLettered()
	default:
		b.metrics.IncQueueDeadLettered()
	}
}

func (b *Broker) Stats() map[string]int {
	return map[string]int{
		"queued":         len(b.queue),
		"queue_cap":      cap(b.queue),
		"dead_letters":   len(b.dlq),
		"deadletter_cap": cap(b.dlq),
	}
}

func (b *Broker) DeadLetters() []Message {
	size := len(b.dlq)
	out := make([]Message, 0, size)
	for i := 0; i < size; i++ {
		select {
		case msg := <-b.dlq:
			out = append(out, msg)
			b.dlq <- msg
		default:
			return out
		}
	}
	return out
}

func (b *Broker) Close() {
	b.once.Do(func() {
		b.closed.Store(true)
		close(b.queue)
	})
}
