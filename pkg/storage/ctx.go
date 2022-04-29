package storage

import "context"

type storeContextKey struct {
}

func StoreFromContext(ctx context.Context) Store {
	return ctx.Value(storeContextKey{}).(Store)
}

func NewContextWithStore(ctx context.Context, s Store) context.Context {
	return context.WithValue(ctx, storeContextKey{}, s)
}
