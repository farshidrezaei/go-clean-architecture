package memory

import "context"

type UnitOfWork struct {
	store *Store
}

func NewUnitOfWork(store *Store) *UnitOfWork {
	return &UnitOfWork{store: store}
}

func (u *UnitOfWork) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	u.store.mu.Lock()
	defer u.store.mu.Unlock()

	snap := u.store.cloneLocked()
	if err := fn(withSnapshot(ctx, snap)); err != nil {
		return err
	}

	u.store.commitLocked(snap)
	return nil
}
