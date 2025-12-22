package logging

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	originalsentry "github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
)

type TransportMock struct {
	mu        sync.Mutex
	events    []*originalsentry.Event
	lastEvent *originalsentry.Event
}

var (
	_ originalsentry.Transport = (*TransportMock)(nil)
)

func (t *TransportMock) Configure(options originalsentry.ClientOptions) {}
func (t *TransportMock) SendEvent(event *originalsentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
	t.lastEvent = event
}
func (t *TransportMock) Flush(timeout time.Duration) bool {
	return true
}
func (t *TransportMock) FlushWithContext(ctx context.Context) bool {
	return true
}
func (t *TransportMock) Events() []*originalsentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.events
}
func (t *TransportMock) Close() {}

func TestSentry(t *testing.T) {
	require := require.New(t)
	transport := &TransportMock{}
	defer func() {
		require.Len(transport.events, 1)
	}()
	conf := SentryConfig{
		Level:     "ERROR",
		Transport: transport,
	}
	h := NewSentryHandler(&conf, "test")
	defer h.Close()
	log := slog.New(h)
	slog.SetDefault(log)
	slog.Error("This is test")
}
