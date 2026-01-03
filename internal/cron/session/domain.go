package session

import (
	"time"
)

// Activity represents session activity data from Redis
type Activity struct {
	SessionID    string
	LastActivity time.Time
}

func (s *Activity) IsExpired() bool {
	return s.LastActivity.Add(time.Minute * 5).Before(time.Now())
}
