package lib

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

type ULIDPrefix string

const (
	ULIDRequestIDPrefix ULIDPrefix = "req"
	ULIDGameIDPrefix    ULIDPrefix = "game"
)

// NewULID generates a new ULID
func NewULID(prefix ULIDPrefix) string {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	return fmt.Sprintf("%s-%s", prefix, ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String())
}

// NewRequestID generates a new request ID using ULID
func NewRequestID() string {
	return NewULID(ULIDRequestIDPrefix)
}
