package collections

import (
	"time"
)

type Order struct {
	HotelID        string
	RoomCategoryID string
	UserEmail      string
	DateFrom       time.Time
	DateTo         time.Time
}
