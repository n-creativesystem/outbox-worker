package logging

import (
	"context"
	"io"
	"log/slog"
	"strconv"

	"go.opentelemetry.io/otel/trace"
)

type Service interface {
	ServiceName() string
	Environment() string
	Version() string
}

type datadogHandler struct {
	service Service
	slog.Handler
}

func NewDatadogHandler(service Service, handler slog.Handler) Handle {
	return &datadogHandler{
		service: service,
		Handler: handler,
	}
}

func (h *datadogHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		// e.g. https://docs.datadoghq.com/ja/tracing/other_telemetry/connect_logs_and_traces/opentelemetry/?tab=go
		record.AddAttrs(slog.String("dd.trace_id", convertTraceID(span.SpanContext().TraceID().String())))
		record.AddAttrs(slog.String("dd.span_id", convertTraceID(span.SpanContext().SpanID().String())))
		record.AddAttrs(slog.String("dd.service", h.service.ServiceName()))
		record.AddAttrs(slog.String("dd.env", h.service.Environment()))
		record.AddAttrs(slog.String("dd.version", h.service.Version()))
	}
	return h.Handler.Handle(ctx, record)
}

func (h *datadogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewDatadogHandler(h.service, h.Handler.WithAttrs(attrs))
}

func (h *datadogHandler) WithGroup(name string) slog.Handler {
	return NewDatadogHandler(h.service, h.Handler.WithGroup(name))
}

func (h *datadogHandler) Close() error {
	if v, ok := h.Handler.(io.Closer); ok {
		_ = v.Close() // nolint: errcheck
	}
	return nil
}

func convertTraceID(id string) string {
	if len(id) < 16 {
		return ""
	}
	if len(id) > 16 {
		id = id[16:]
	}
	intValue, err := strconv.ParseUint(id, 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatUint(intValue, 10)
}
