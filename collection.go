package in_memory

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"
)

const (
	emptyOwnerID uint64 = 0
)

var (
	ErrEmptyElemsList            = errors.New("elements list in method argument cannot be nil or empty")
	ErrEmptyModifier             = errors.New("modifier in method argument cannot be nil")
	ErrElemNotFound              = errors.New("queried element not found")
	ErrNotAllowedOutsideExecutor = errors.New("queries are not allowed outside executor")
)

type ownerStruct struct {
	id atomic.Uint64
	_  [56]byte // Alignment: на будущее. Ради выравнивания в рамках линии кеша процессора. Можно попробовать поиграться с шардированием атомика
	// Но в данном кейсе не придумал как сделать (source https://www.youtube.com/watch?app=desktop&v=TjzeCWaTOtM&t=0s)
}

type Collection[T any] struct {
	name  string
	count atomic.Uint64
	owner ownerStruct

	data map[uint64]T
}

func newCollection[T any](name string) Collection[T] {
	return Collection[T]{
		name:  name,
		owner: ownerStruct{},
		data:  make(map[uint64]T),
	}
}

func (c *Collection[T]) TryOwn(ctx context.Context, ownerID uint64) (bool, error) {
	if !isQueryAllowed(ctx) {
		return false, ErrNotAllowedOutsideExecutor
	}

	return c.owner.id.CompareAndSwap(emptyOwnerID, ownerID), nil
}

func (c *Collection[T]) ResetOwner(ctx context.Context, ownerID uint64) error {
	if !isQueryAllowed(ctx) {
		return ErrNotAllowedOutsideExecutor
	}

	_ = c.owner.id.CompareAndSwap(ownerID, emptyOwnerID)

	return nil
}

func (c *Collection[T]) SelectByIDs(ctx context.Context, ids []uint64) ([]T, error) {
	if !isQueryAllowed(ctx) {
		return nil, ErrNotAllowedOutsideExecutor
	}

	var result []T

	for i, v := range c.data {
		if slices.Contains(ids, i) {
			result = append(result, v)
		}
	}

	return result, nil
}

func (c *Collection[T]) Select(ctx context.Context, filter func(elem T) bool) ([]T, error) {
	if !isQueryAllowed(ctx) {
		return nil, ErrNotAllowedOutsideExecutor
	}

	var result []T

	for _, v := range c.data {
		if filter(v) {
			result = append(result, v)
		}
	}

	return result, nil
}

func (c *Collection[T]) Insert(ctx context.Context, elems []T) ([]uint64, error) {
	if !isQueryAllowed(ctx) {
		return nil, ErrNotAllowedOutsideExecutor
	}

	if len(elems) == 0 {
		return nil, ErrEmptyElemsList
	}

	resultIDs := make([]uint64, 0, len(elems))
	for _, elem := range elems {
		newID := c.count.Add(uint64(1))

		c.data[newID] = elem

		resultIDs = append(resultIDs, newID)
	}

	return resultIDs, nil
}

func (c *Collection[T]) UpdateByIDs(ctx context.Context, ids []uint64, modifier func(elem T) T) error {
	if !isQueryAllowed(ctx) {
		return ErrNotAllowedOutsideExecutor
	}

	if len(ids) == 0 {
		return ErrEmptyElemsList
	}

	if modifier == nil {
		return ErrEmptyModifier
	}

	for _, id := range ids {
		if _, ok := c.data[id]; !ok {
			return ErrElemNotFound
		}
	}

	for _, id := range ids {
		c.data[id] = modifier(c.data[id])
	}

	return nil
}

func (c *Collection[T]) Delete(ctx context.Context, ids []uint64) error {
	if !isQueryAllowed(ctx) {
		return ErrNotAllowedOutsideExecutor
	}

	if len(ids) == 0 {
		return ErrEmptyElemsList
	}

	for _, id := range ids {
		delete(c.data, id)
	}

	return nil
}

func isQueryAllowed(ctx context.Context) bool {
	return contextGetOwnerID(ctx) > 0
}
