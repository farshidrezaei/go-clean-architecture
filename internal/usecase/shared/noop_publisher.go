package shared

import "context"

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, Event) {}
