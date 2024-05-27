package collections

import (
	"time"
)

type RoomAvailability struct {
	HotelID string
	RoomID  string
	Date    time.Time
	Quota   int
}
