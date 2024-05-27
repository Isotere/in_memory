package in_memory

import (
	"context"
	"runtime"
)

type TransactionBodyFunc func(ctx context.Context, collectionsList *CollectionsList) error

type locker interface {
	TryOwn(ctx context.Context, ownerID uint64) (bool, error)
	ResetOwner(ctx context.Context, ownerID uint64) error
}

func (c *InMemoryConnection) Execute(ctx context.Context, lockerList []locker, fArgs ...TransactionBodyFunc) error {
	// Чтобы запрос не завис на очень долго, ставим таймаут (из настроек соединения)
	ctx, cancel := context.WithTimeout(ctx, c.timeoutMs)
	defer cancel()

	ownerID := c.GenerateQueryID()

	// Мы не должны позволять запускать прямые запросы из-вне экзекютора, поэтому пропишем в контексте, а в запросах будем проверять
	ctx = contextSetOwnerID(ctx, ownerID)

	for {
		// Проверимся на таймаут
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Need success ALL collections owned
		if tryOwn(ctx, ownerID, lockerList) {
			break
		}

		// Если не получилось - сбросим все, что успели заовнить
		resetOwner(ctx, ownerID, lockerList)

		runtime.Gosched()
	}

	// На выходе из функции сбросим все блокировки которые были
	defer resetOwner(ctx, ownerID, lockerList)

	for _, f := range fArgs {
		err := f(ctx, c.Collections())
		if err != nil {
			return err
		}
	}

	return nil
}

func tryOwn(ctx context.Context, queryID uint64, lockerList []locker) bool {
	for _, current := range lockerList {
		// Сознательно игнорирую возможную ошибку доступа (ее можно там залогировать, но тут думаю достаточно вернуть false)
		if ok, _ := current.TryOwn(ctx, queryID); !ok {
			return false
		}
	}

	return true
}

func resetOwner(ctx context.Context, queryID uint64, lockerList []locker) {
	for _, current := range lockerList {
		_ = current.ResetOwner(ctx, queryID)
	}
}
