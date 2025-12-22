package outbox

import (
	"context"
	"testing"
	"time"

	"github.com/n-creativesystem/outbox-worker/pkg/domain/outbox"
	"github.com/stretchr/testify/require"
)

func TestClient_PingContext(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			err := c.PingContext(context.Background())
			require.NoError(t, err)
		})
	}
}

func TestClient_PingContext_WithClosedDB(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			// Close the database
			err := c.db.Close()
			require.NoError(t, err)

			// Ping should fail
			err = c.PingContext(context.Background())
			require.Error(t, err)
		})
	}
}

func TestClient_FindEvents_BuildEventModels(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			now := time.Date(2025, 12, 21, 10, 0, 0, 0, time.UTC)
			retryAt := now.Add(10 * time.Minute)
			sentAt := now.Add(-1 * time.Minute)

			// Insert rows. (sent_at is NULL so it is publishable)
			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", `{"a":1}`, nil, nil, nil},
				{"orders", "A2", "updated", `{"a":2}`, retryAt, 3, nil},
				{"payments", "P1", "paid", `{"a":3}`, retryAt, 1, sentAt},
			})

			ctx := context.Background()
			events, err := c.FindEvents(ctx, nil, nil, nil)
			require.NoError(t, err)
			require.Len(t, events, 3)

			require.Equal(t, int64(1), events[0].ID())
			require.Equal(t, "orders", events[0].AggregateType())
			require.Equal(t, "A1", events[0].AggregateID())
			require.Equal(t, "created", events[0].Event())
			require.Equal(t, "test-producer", events[0].ProducerName())
			require.JSONEq(t, `{"a":1}`, events[0].Payload())
			require.Nil(t, events[0].RetryAt())
			require.Equal(t, 0, events[0].RetryCount())

			require.Equal(t, int64(2), events[1].ID())
			require.NotNil(t, events[1].RetryAt())
			require.Equal(t, retryAt.Unix(), events[1].RetryAt().Unix())
			require.Equal(t, 3, events[1].RetryCount())

			require.Equal(t, int64(3), events[2].ID())
			require.NotNil(t, events[2].RetryAt())
			require.Equal(t, retryAt.Unix(), events[2].RetryAt().Unix())
		})
	}
}

func TestClient_FindEvents_Filtering_MaxRetryCount(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
				{"orders", "A2", "updated", "{}", nil, 9, nil},
				{"orders", "A3", "updated", "{}", nil, 10, nil},
				{"orders", "A4", "updated", "{}", nil, 11, nil},
			})

			ctx := context.Background()
			events, err := c.FindEvents(ctx, nil, nil, nil)
			require.NoError(t, err)
			// MaxRetryCount is 10, so only events with retry_count < 10 should be returned
			require.Len(t, events, 2)
			require.Equal(t, int64(1), events[0].ID())
			require.Equal(t, int64(2), events[1].ID())
		})
	}
}

func TestClient_FindEvents_Filtering_WhitelistResources(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
				{"payments", "P1", "paid", "{}", nil, nil, nil},
				{"inventory", "I1", "restocked", "{}", nil, nil, nil},
			})

			ctx := context.Background()
			// Only get orders and payments
			events, err := c.FindEvents(ctx, []string{"orders", "payments"}, nil, nil)
			require.NoError(t, err)
			require.Len(t, events, 2)
			require.Equal(t, int64(1), events[0].ID())
			require.Equal(t, "orders", events[0].AggregateType())
			require.Equal(t, int64(2), events[1].ID())
			require.Equal(t, "payments", events[1].AggregateType())
		})
	}
}

func TestClient_FindEvents_Filtering_BlacklistResources(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
				{"payments", "P1", "paid", "{}", nil, nil, nil},
				{"inventory", "I1", "restocked", "{}", nil, nil, nil},
			})

			ctx := context.Background()
			// Exclude payments
			events, err := c.FindEvents(ctx, nil, []string{"payments"}, nil)
			require.NoError(t, err)
			require.Len(t, events, 2)
			require.Equal(t, int64(1), events[0].ID())
			require.Equal(t, "orders", events[0].AggregateType())
			require.Equal(t, int64(3), events[1].ID())
			require.Equal(t, "inventory", events[1].AggregateType())
		})
	}
}

func TestClient_FindEvents_Filtering_SkipEvents(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
				{"orders", "A2", "updated", "{}", nil, nil, nil},
				{"orders", "A3", "deleted", "{}", nil, nil, nil},
			})

			ctx := context.Background()
			// Skip deleted events
			events, err := c.FindEvents(ctx, nil, nil, []string{"deleted"})
			require.NoError(t, err)
			require.Len(t, events, 2)
			require.Equal(t, int64(1), events[0].ID())
			require.Equal(t, "created", events[0].Event())
			require.Equal(t, int64(2), events[1].ID())
			require.Equal(t, "updated", events[1].Event())
		})
	}
}

func TestClient_FindEvents_Filtering_Combined(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			// t.Parallel()

			c := newTestClientForProvider(t, provider)

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
				{"payments", "P1", "paid", "{}", nil, nil, nil},
				{"orders", "A2", "deleted", "{}", nil, nil, nil},
				{"inventory", "I1", "restocked", "{}", nil, nil, nil},
				{"orders", "A3", "updated", "{}", nil, nil, nil},
			})

			ctx := context.Background()
			// Whitelist: orders, payments
			// Blacklist: payments
			// Skip: deleted
			// Expected: orders (created, updated)
			events, err := c.FindEvents(ctx, []string{"orders", "payments"}, []string{"payments"}, []string{"deleted"})
			require.NoError(t, err)
			require.Len(t, events, 2)
			require.Equal(t, int64(1), events[0].ID())
			require.Equal(t, "orders", events[0].AggregateType())
			require.Equal(t, "created", events[0].Event())
			require.Equal(t, int64(5), events[1].ID())
			require.Equal(t, "orders", events[1].AggregateType())
			require.Equal(t, "updated", events[1].Event())
		})
	}
}

func TestClient_Transaction_isOk_updates_sent_at(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			baseNow := time.Date(2025, 12, 21, 10, 0, 0, 0, time.UTC)
			c.timeNow = func() time.Time { return baseNow }

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
			})

			evt := outbox.NewEvent(1, "orders", "A1", "created", "producer", "{}", nil, 0, nil)
			require.NoError(t, c.Transaction(context.Background(), true, evt))

			var sentAt *time.Time
			err := c.db.GetContext(context.Background(), &sentAt, `SELECT sent_at FROM outbox WHERE id=1`)
			require.NoError(t, err)
			require.NotNil(t, sentAt)
			require.Equal(t, baseNow.Unix(), sentAt.Unix())
		})
	}
}

func TestClient_Transaction_isNotOk_updates_retry_fields(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			baseNow := time.Date(2025, 12, 21, 10, 0, 0, 0, time.UTC)
			retryAt := baseNow.Add(30 * time.Second)

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
			})

			evt := outbox.NewEvent(1, "orders", "A1", "created", "producer", "{}", &retryAt, 7, nil)
			require.NoError(t, c.Transaction(context.Background(), false, evt))

			var gotRetryAt time.Time
			var gotRetryCount int
			err := c.db.QueryRowContext(context.Background(), `SELECT retry_at, retry_count FROM outbox WHERE id=1`).Scan(&gotRetryAt, &gotRetryCount)
			require.NoError(t, err)
			require.Equal(t, retryAt.Unix(), gotRetryAt.Unix())
			require.Equal(t, 7, gotRetryCount)
		})
	}
}

func TestClient_Transaction_Rollback_OnError(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			insertTestData(t, provider, c, [][]any{
				{"orders", "A1", "created", "{}", nil, nil, nil},
			})

			// Create an event with invalid ID that doesn't exist
			evt := outbox.NewEvent(999, "orders", "A1", "created", "producer", "{}", nil, 0, nil)
			err := c.Transaction(context.Background(), true, evt)
			// Transaction should succeed even if no rows are affected
			require.NoError(t, err)
		})
	}
}

func TestClient_outboxTableExpressions_EmptyFilters(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			expr := c.outboxTableExpressions(nil, nil, nil)
			require.Nil(t, expr)
		})
	}
}

func TestClient_outboxTableExpressions_OnlyWhitelist(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			expr := c.outboxTableExpressions([]string{"orders", "payments"}, nil, nil)
			require.NotNil(t, expr)
		})
	}
}

func TestClient_outboxTableExpressions_OnlyBlacklist(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			expr := c.outboxTableExpressions(nil, []string{"payments"}, nil)
			require.NotNil(t, expr)
		})
	}
}

func TestClient_outboxTableExpressions_OnlySkipEvents(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			expr := c.outboxTableExpressions(nil, nil, []string{"deleted"})
			require.NotNil(t, expr)
		})
	}
}

func TestClient_outboxTableExpressions_AllFilters(t *testing.T) {
	t.Parallel()

	for _, provider := range testProviders() {
		t.Run(provider.Name(), func(t *testing.T) {
			t.Parallel()

			c := newTestClientForProvider(t, provider)

			expr := c.outboxTableExpressions(
				[]string{"orders", "payments"},
				[]string{"payments"},
				[]string{"deleted"},
			)
			require.NotNil(t, expr)
		})
	}
}
