package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/antsurge/weaver-admin/pkg/utils/message"
)

func collect(t *testing.T, count int, timeout time.Duration) (func(context.Context, *message.Message) error, func() []string) {
	t.Helper()
	var mu sync.Mutex
	got := make([]string, 0, count)

	h := func(_ context.Context, m *message.Message) error {
		mu.Lock()
		got = append(got, string(m.Body))
		mu.Unlock()
		return nil
	}
	read := func() []string {
		mu.Lock()
		defer mu.Unlock()
		out := make([]string, len(got))
		copy(out, got)
		return out
	}
	return h, read
}

func TestBroadcastEverySubscriberReceives(t *testing.T) {
	b, err := New(&message.Config{Driver: DriverName, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	h1, read1 := collect(t, 1, time.Second)
	h2, read2 := collect(t, 1, time.Second)

	if err := b.Subscribe(context.Background(), "t", message.SubscribeOption{}, h1); err != nil {
		t.Fatal(err)
	}
	if err := b.Subscribe(context.Background(), "t", message.SubscribeOption{}, h2); err != nil {
		t.Fatal(err)
	}

	if err := b.Publish(context.Background(), message.NewMessage("t", []byte("hello"))); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(read1()) > 0 && len(read2()) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := read1(); len(got) != 1 || got[0] != "hello" {
		t.Fatalf("subscriber1 got %v, want [hello]", got)
	}
	if got := read2(); len(got) != 1 || got[0] != "hello" {
		t.Fatalf("subscriber2 got %v, want [hello]", got)
	}
}

func TestGroupOnlyOneSubscriberReceives(t *testing.T) {
	b, err := New(&message.Config{Driver: DriverName, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	h1, read1 := collect(t, 1, time.Second)
	h2, read2 := collect(t, 1, time.Second)

	opt := message.SubscribeOption{Group: "g1"}
	if err := b.Subscribe(context.Background(), "t", opt, h1); err != nil {
		t.Fatal(err)
	}
	if err := b.Subscribe(context.Background(), "t", opt, h2); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := b.Publish(context.Background(), message.NewMessage("t", []byte("x"))); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(300 * time.Millisecond)
	total := len(read1()) + len(read2())
	if total != 10 {
		t.Fatalf("group total = %d (s1=%d s2=%d), want 10", total, len(read1()), len(read2()))
	}
}

func TestContextCancelUnsubscribes(t *testing.T) {
	b, err := New(&message.Config{Driver: DriverName, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	h, read := collect(t, 1, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	if err := b.Subscribe(ctx, "t", message.SubscribeOption{}, h); err != nil {
		t.Fatal(err)
	}
	cancel()
	time.Sleep(100 * time.Millisecond)

	if err := b.Publish(context.Background(), message.NewMessage("t", []byte("late"))); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if got := read(); len(got) != 0 {
		t.Fatalf("canceled subscriber still received %v", got)
	}
}
