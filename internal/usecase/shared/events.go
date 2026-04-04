package shared

import "context"

type Event struct {
	Name     string
	ActorID  string
	TargetID string
	Metadata map[string]string
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event)
}
