package logging

import (
	"context"
	"log/slog"
)

const (
	TxKey       = "tx_id"
	IdentityKey = "identity_id"
)

type ContextHandler struct {
	Handler slog.Handler
}

func (h ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if txId, ok := ctx.Value(TxKey).(string); ok {
		r.AddAttrs(slog.String(TxKey, txId))
	}

	if identityId, ok := ctx.Value(IdentityKey).(string); ok {
		r.AddAttrs(slog.String(IdentityKey, identityId))
	}

	return h.Handler.Handle(ctx, r)
}

func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.Handler.WithAttrs(attrs)
}

func (h ContextHandler) WithGroup(name string) slog.Handler {
	return h.Handler.WithGroup(name)
}
