package pagination

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

const DefaultLimit = 20
const MaxLimit = 100

// Cursor encodes a (created_at, id) pair for keyset pagination.
// Opaque to the client — never parse or depend on the format externally.
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

func Encode(createdAt time.Time, id string) string {
	raw := fmt.Sprintf("%d|%s", createdAt.UnixNano(), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func Decode(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}
	b, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("pagination: invalid cursor encoding")
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("pagination: malformed cursor")
	}
	var nanos int64
	if _, err := fmt.Sscan(parts[0], &nanos); err != nil {
		return nil, fmt.Errorf("pagination: cursor timestamp invalid")
	}
	return &Cursor{
		CreatedAt: time.Unix(0, nanos).UTC(),
		ID:        parts[1],
	}, nil
}

// ClampLimit returns limit clamped to [1, MaxLimit], defaulting to DefaultLimit.
func ClampLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}
