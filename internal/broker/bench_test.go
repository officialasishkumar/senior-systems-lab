package broker

import (
	"context"
	"testing"

	"github.com/officialasishkumar/senior-systems-lab/internal/observability"
)

func BenchmarkPublish(b *testing.B) {
	broker := New(b.N, 1, observability.NewMetrics())
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := broker.Publish(ctx, Message{Topic: "bench", Payload: "payload"}); err != nil {
			b.Fatal(err)
		}
	}
}
