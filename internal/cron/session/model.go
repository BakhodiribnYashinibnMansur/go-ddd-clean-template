package session

import "time"

// SessionActivity represents session activity data from Redis
type SessionActivity struct {
	SessionID    string
	LastActivity time.Time
}
