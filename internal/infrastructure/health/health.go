package health

import "context"

type Checker interface {
	Check(ctx context.Context) error
}

type MemoryChecker struct{}

func (MemoryChecker) Check(context.Context) error {
	return nil
}
