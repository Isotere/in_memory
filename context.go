package in_memory

import "context"

type (
	keyOwnerID struct{}
)

func contextGetOwnerID(ctx context.Context) uint64 {
	val := ctx.Value(keyOwnerID{})

	if v, ok := val.(uint64); ok {
		return v
	}

	return 0
}

func contextSetOwnerID(ctx context.Context, ownerID uint64) context.Context {
	return context.WithValue(ctx, keyOwnerID{}, ownerID)
}
