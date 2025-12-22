package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/jmoiron/sqlx"
	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/domain/outbox"
	"github.com/n-creativesystem/outbox-worker/pkg/infrastructure/repository/outbox/schema"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/multierr"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	postgres "github.com/doug-martin/goqu/v9/dialect/postgres"
)

func init() {
	goqu.RegisterDialect("pgx", postgres.DialectOptions())
}

type Client struct {
	db      *sqlx.DB
	cfg     *config.OutboxPolling
	timeNow func() time.Time
}

var _ outbox.Repository = (*Client)(nil)

func NewClient(ctx context.Context, cfg *config.OutboxPolling, opts ...ClientOption) (*Client, error) {
	dsn, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("rdb: failed to build connection string: %w", err)
	}
	dialect := cfg.Database.Dialect()
	db, err := otelsql.Open(dialect, dsn)
	if err != nil {
		return nil, err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(timeoutCtx); err != nil {
		return nil, err
	}
	c := newClient(ctx, cfg, dialect, db)
	return c, nil
}

func newClient(_ context.Context, cfg *config.OutboxPolling, dialect string, db *sql.DB, opts ...ClientOption) *Client {
	c := &Client{
		db:      sqlx.NewDb(db, dialect),
		cfg:     cfg,
		timeNow: time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) PingContext(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *Client) FindEvents(
	ctx context.Context,
	whitelistResources []string,
	blacklistResources []string,
	skipEvents []string,
) (_ []*outbox.Event, rErr error) {
	ctx = trace.StartSpan(ctx, "findEvents")
	defer func() { trace.EndSpan(ctx, rErr) }()

	expressions := []exp.Expression{
		goqu.Or(
			goqu.C("retry_count").IsNull(),
			goqu.C("retry_count").Lt(c.cfg.OutboxConfig.MaxRetryCount),
		),
	}
	if expr := c.outboxTableExpressions(whitelistResources, blacklistResources, skipEvents); expr != nil {
		expressions = append(expressions, expr)
	}

	sql, args, err := goqu.Dialect(c.cfg.Database.Dialect()).
		Select(&schema.Outbox{}).
		From(c.cfg.OutboxConfig.Table()).
		Where(expressions...).
		Order(goqu.I("id").Asc()).
		Limit(c.cfg.OutboxConfig.FindEventLimit).
		Prepared(true).
		ToSQL()
	if err != nil {
		return nil, err
	}
	var events []schema.Outbox
	err = c.db.SelectContext(ctx, &events, sql, args...)
	if err != nil {
		return nil, err
	}

	results := make(outbox.Events, len(events))
	for idx := range events {
		event := events[idx]
		results[idx] = outbox.NewEvent(
			event.ID,
			event.AggregateType,
			event.AggregateID,
			event.Event,
			c.cfg.OutboxConfig.ProducerName,
			string(event.Payload),
			event.RetryAtValue(),
			event.RetryCountValue(),
			event.SentAtValue(),
		)
	}
	return results, nil
}

func (c *Client) Transaction(ctx context.Context, isOk bool, event *outbox.Event) (err error) {
	defer func() {
		if pErr := recover(); pErr != nil {
			slog.Error(fmt.Sprintf("recover from panic: %+v", pErr))
		}
	}()
	tx, err := c.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = multierr.Append(err, txErr)
			}
		} else {
			if txErr := tx.Commit(); txErr != nil {
				err = txErr
			}
		}
	}()
	if isOk {
		return c.txUpdateSentAtField(ctx, tx, event)
	} else {
		return c.txUpdateRetryField(ctx, tx, event)
	}
}

func (c *Client) txUpdateSentAtField(ctx context.Context, tx *sqlx.Tx, event *outbox.Event) (rErr error) {
	ctx = trace.StartSpan(ctx, "txUpdateSentAtRecord",
		attribute.Int64("ID", event.ID()),
	)
	defer func() { trace.EndSpan(ctx, rErr) }()

	ds := goqu.Dialect(c.cfg.Database.Dialect()).
		Update(c.cfg.OutboxConfig.Table()).
		Set(
			goqu.Record{
				"sent_at": c.timeNow(),
			},
		)
	sql, args, err := ds.
		Where(
			goqu.Ex{
				"id": event.ID(),
			},
		).
		Prepared(true).
		ToSQL()
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return err
	}
	slog.With(slog.Int64("ID", event.ID())).InfoContext(ctx, "Published and update sent_at field.")
	return nil
}

func (c *Client) txUpdateRetryField(ctx context.Context, tx *sqlx.Tx, event *outbox.Event) (rErr error) {
	ctx = trace.StartSpan(ctx, "txUpdateRetryField",
		attribute.Int64("ID", event.ID()),
	)
	defer func() { trace.EndSpan(ctx, rErr) }()

	sql, args, err := goqu.Dialect(c.cfg.Database.Dialect()).
		Update(c.cfg.OutboxConfig.Table()).
		Prepared(true).
		Set(
			goqu.Record{
				"retry_count": event.RetryCount(),
				"retry_at":    *event.RetryAt(),
			},
		).
		Where(
			goqu.Ex{
				"id": event.ID(),
			},
		).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, sql, args...)
	slog.With(
		slog.Int64("ID", event.ID()),
		slog.Int("retry_count", event.RetryCount()),
		slog.Time("retry_at", *event.RetryAt()),
	).WarnContext(ctx, "Update retry column.")
	return err
}

func (c *Client) outboxTableExpressions(
	whitelistResources []string,
	blacklistResources []string,
	skipEvents []string,
) exp.Expression {
	resourceExpr := []exp.Expression{}
	if v := whitelistResources; len(v) > 0 {
		resourceExpr = append(resourceExpr, goqu.C("aggregate_type").In(v))
	}
	if v := blacklistResources; len(v) > 0 {
		resourceExpr = append(resourceExpr, goqu.C("aggregate_type").NotIn(v))
	}
	var expressions exp.Expression
	if len(resourceExpr) > 0 {
		expressions = goqu.And(resourceExpr...)
	}
	if v := skipEvents; len(v) > 0 {
		if expressions != nil {
			return goqu.And(expressions, goqu.C("event").NotIn(v))
		} else {
			return goqu.C("event").NotIn(v)
		}
	}
	return expressions
}
