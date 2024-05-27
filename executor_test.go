package in_memory

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Isotere/in_memory/collections"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	ctx := context.Background()
	connection := NewConnection(700)

	procs := runtime.GOMAXPROCS(0) * 15
	fmt.Printf("GOMAXPROCS %d\n", procs)

	err := testInsertBatchRoomsAvailable(ctx, connection)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(procs)
	for i := 0; i < procs; i++ {
		go func() {
			defer wg.Done()

			errRoutine := connection.Execute(
				ctx,
				[]locker{
					&connection.Collections().Orders,
					&connection.Collections().RoomAvailability,
				},
				testInsertOrder,
				testDecreaseRoomsAvailable,
			)
			assert.NoError(t, errRoutine)
		}()
	}

	wg.Wait()

	// Чтобы тут не заморачиваться с экзекутором
	ctx = contextSetOwnerID(ctx, testOwnerID)

	res0, _ := connection.Collections().Orders.Select(ctx, func(elem collections.Order) bool {
		return len(elem.RoomCategoryID) > 0
	})
	assert.NotNil(t, res0)
	assert.Len(t, res0, 150)

	res, _ := connection.Collections().RoomAvailability.SelectByIDs(ctx, []uint64{1})
	assert.NotNil(t, res)
	assert.Equal(t, 150, math.MaxInt-res[0].Quota)
}

func testInsertBatchRoomsAvailable(ctx context.Context, conn *InMemoryConnection) error {
	return conn.Execute(
		ctx,
		[]locker{
			&conn.Collections().RoomAvailability,
		},
		func(ctx context.Context, collectionsList *CollectionsList) error {
			_, errInner := collectionsList.RoomAvailability.Insert(ctx, []collections.RoomAvailability{{
				HotelID: "raddison",
				RoomID:  "lux",
				Date:    time.Now(),
				Quota:   math.MaxInt,
			}})

			return errInner
		},
	)
}

func testInsertOrder(ctx context.Context, collectionsList *CollectionsList) error {
	r := rand.N[int](5)
	time.Sleep(time.Millisecond * time.Duration(r))

	_, err := collectionsList.Orders.Insert(ctx, []collections.Order{{
		HotelID:        "raddison",
		RoomCategoryID: "lux",
		UserEmail:      "admin@example.com",
		DateFrom:       time.Now(),
		DateTo:         time.Now(),
	}})

	return err
}

func testDecreaseRoomsAvailable(ctx context.Context, collectionsList *CollectionsList) error {
	r := rand.N[int](5)
	time.Sleep(time.Millisecond * time.Duration(r))

	return collectionsList.RoomAvailability.UpdateByIDs(ctx, []uint64{1}, func(elem collections.RoomAvailability) collections.RoomAvailability {
		elem.Quota--
		return elem
	})
}
