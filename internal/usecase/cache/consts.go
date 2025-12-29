package cache

import (
	"errors"
	"time"
)

var ErrNilOutput = errors.New("out must not be nil")

const (
	publicCacheTime = 1 * time.Hour
)
