package in_memory

import (
	"context"
	"testing"
	"time"

	"github.com/Isotere/in_memory/collections"
	"github.com/stretchr/testify/assert"
)

const (
	testOwnerID  uint64 = 1
	testOwnerID2 uint64 = 2
)

func TestTryOwn(t *testing.T) {
	ctx := contextSetOwnerID(context.Background(), testOwnerID)

	orderCollection := newCollection[collections.Order]("order")

	res, err := orderCollection.tryOwn(ctx, testOwnerID)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = orderCollection.tryOwn(ctx, testOwnerID2)
	assert.NoError(t, err)
	assert.False(t, res)
}

func TestResetOwner(t *testing.T) {
	ctx := contextSetOwnerID(context.Background(), testOwnerID)

	orderCollection := newCollection[collections.Order]("order")

	res, err := orderCollection.tryOwn(ctx, testOwnerID)
	assert.NoError(t, err)
	assert.True(t, res)
	assert.Equal(t, testOwnerID, orderCollection.owner.id.Load())

	err = orderCollection.resetOwner(ctx, testOwnerID)
	assert.NoError(t, err)
	assert.Equal(t, emptyOwnerID, orderCollection.owner.id.Load())
}

func TestSelectByIDs(t *testing.T) {
	ctx := contextSetOwnerID(context.Background(), testOwnerID)

	t.Run("empty collection", func(t *testing.T) {
		orderCollection := newCollection[collections.Order]("order")

		res, err := orderCollection.SelectByIDs(ctx, []ObjectID{1, 2})
		assert.NoError(t, err)

		assert.Nil(t, res)
	})

	t.Run("not empty collection, but no result", func(t *testing.T) {
		orderCollection := newCollection[collections.Order]("order")
		orderCollection.data[ObjectID(1)] = collections.Order{}
		orderCollection.data[ObjectID(2)] = collections.Order{}

		res, err := orderCollection.SelectByIDs(ctx, []ObjectID{3, 4})
		assert.NoError(t, err)

		assert.Nil(t, res)
	})

	t.Run("not empty collection, with result", func(t *testing.T) {
		orderCollection := newCollection[collections.Order]("order")
		orderCollection.data[ObjectID(1)] = collections.Order{}
		orderCollection.data[ObjectID(2)] = collections.Order{}
		orderCollection.data[ObjectID(3)] = collections.Order{}
		orderCollection.data[ObjectID(4)] = collections.Order{}

		res, err := orderCollection.SelectByIDs(ctx, []ObjectID{2, 3, 5})
		assert.NoError(t, err)

		assert.Len(t, res, 2)
	})
}

func TestSelect(t *testing.T) {
	ctx := contextSetOwnerID(context.Background(), testOwnerID)

	orderCollection := testGetDummyOrderCollection()

	t.Run("empty result", func(t *testing.T) {
		res, err := orderCollection.Select(ctx, func(elem collections.Order) bool {
			return elem.UserEmail == "some@some.com"
		})
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("not empty result", func(t *testing.T) {
		res, err := orderCollection.Select(ctx, func(elem collections.Order) bool {
			return elem.UserEmail == "admin@email.ru"
		})
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})
}

func TestInsert(t *testing.T) {
	ctx := contextSetOwnerID(context.Background(), testOwnerID)

	orderCollection := testGetDummyOrderCollection()

	t.Run("empty args", func(t *testing.T) {
		_, err := orderCollection.Insert(ctx, []collections.Order{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyElemsList)
	})

	t.Run("success", func(t *testing.T) {
		ids, err := orderCollection.Insert(ctx, []collections.Order{
			{
				HotelID:        "slav",
				RoomCategoryID: "president",
				UserEmail:      "ya@ya.ru",
				DateFrom:       time.Now(),
				DateTo:         time.Now(),
			},
		})
		assert.NoError(t, err)
		assert.Len(t, ids, 1)

		item := orderCollection.data[ids[0]]
		assert.Equal(t, "president", item.RoomCategoryID)
	})
}

func TestUpdateByIDs(t *testing.T) {
	ctx := contextSetOwnerID(context.Background(), testOwnerID)

	orderCollection := testGetDummyOrderCollection()

	t.Run("empty args", func(t *testing.T) {
		err := orderCollection.UpdateByIDs(ctx, []ObjectID{}, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyElemsList)
	})

	t.Run("nil modifier", func(t *testing.T) {
		err := orderCollection.UpdateByIDs(ctx, []ObjectID{1}, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyModifier)
	})

	t.Run("success", func(t *testing.T) {
		err := orderCollection.UpdateByIDs(ctx, []ObjectID{1}, func(elem collections.Order) collections.Order {
			elem.RoomCategoryID = "mega-lux"

			return elem
		})

		assert.NoError(t, err)

		res, err := orderCollection.SelectByIDs(ctx, []ObjectID{1})
		assert.NoError(t, err)
		assert.Equal(t, "mega-lux", res[ObjectID(1)].RoomCategoryID)
	})
}

func testGetDummyOrderCollection() *Collection[collections.Order] {
	orderCollection := newCollection[collections.Order]("order")
	orderCollection.data[ObjectID(1)] = collections.Order{
		HotelID:        "raddison",
		RoomCategoryID: "lux",
		UserEmail:      "admin@email.ru",
		DateFrom:       time.Time{},
		DateTo:         time.Time{},
	}
	orderCollection.data[ObjectID(2)] = collections.Order{
		HotelID:        "raddison2",
		RoomCategoryID: "lux",
		UserEmail:      "user@email.ru",
		DateFrom:       time.Time{},
		DateTo:         time.Time{},
	}
	orderCollection.data[ObjectID(3)] = collections.Order{
		HotelID:        "raddison3",
		RoomCategoryID: "economy",
		UserEmail:      "admin@email.ru",
		DateFrom:       time.Time{},
		DateTo:         time.Time{},
	}

	return &orderCollection
}
