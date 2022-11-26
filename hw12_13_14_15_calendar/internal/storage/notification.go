package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	EventID    uuid.UUID `json:"eventId"`
	EventTitle string    `json:"eventTitle"`
	DateTime   time.Time `json:"datetime"`
	UserID     int64     `json:"userId"`
}

func (n *Notification) String() string {
	return fmt.Sprintf("New notification from event %s, %s to user with ID %v %s",
		n.EventID.String(),
		n.EventTitle,
		n.UserID,
		n.DateTime.Format("2006-01-02 15:04:05"),
	)
}
