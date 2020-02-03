package ayame

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func getULID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

func unmarshalMessage(c *Connection, rawMessage []byte, v interface{}) error {
	if err := json.Unmarshal(rawMessage, v); err != nil {
		c.trace("invalid JSON, rawMessage: %s, error: %v", rawMessage, err)
		return errorInvalidJSON
	}
	return nil
}

func strPtr(s string) *string {
	return &s
}
