package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"clean_architecture/internal/usecase/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LogPublisher struct {
	logger *slog.Logger
}

func NewLogPublisher(logger *slog.Logger) *LogPublisher {
	return &LogPublisher{logger: logger}
}

func (p *LogPublisher) Publish(_ context.Context, event shared.Event) {
	if p == nil || p.logger == nil {
		return
	}
	attrs := []any{
		"event", event.Name,
		"actor_id", event.ActorID,
		"target_id", event.TargetID,
	}
	for key, value := range event.Metadata {
		attrs = append(attrs, key, value)
	}
	p.logger.Info("audit_event", attrs...)
}

type CompositePublisher struct {
	publishers []shared.EventPublisher
}

func NewCompositePublisher(publishers ...shared.EventPublisher) *CompositePublisher {
	filtered := make([]shared.EventPublisher, 0, len(publishers))
	for _, publisher := range publishers {
		if publisher != nil {
			filtered = append(filtered, publisher)
		}
	}
	return &CompositePublisher{publishers: filtered}
}

func (p *CompositePublisher) Publish(ctx context.Context, event shared.Event) {
	if p == nil {
		return
	}
	for _, publisher := range p.publishers {
		publisher.Publish(ctx, event)
	}
}

type MemoryPublisher struct {
	Events []shared.Event
}

func (p *MemoryPublisher) Publish(_ context.Context, event shared.Event) {
	p.Events = append(p.Events, event)
}

type DBPublisher struct {
	pool *pgxpool.Pool
}

func NewDBPublisher(pool *pgxpool.Pool) *DBPublisher {
	return &DBPublisher{pool: pool}
}

func (p *DBPublisher) Publish(ctx context.Context, event shared.Event) {
	if p == nil || p.pool == nil {
		return
	}
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return
	}
	_, _ = p.pool.Exec(ctx, `
		INSERT INTO audit_logs (id, event_name, actor_id, target_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.NewString(), event.Name, event.ActorID, event.TargetID, metadata, time.Now().UTC())
}
