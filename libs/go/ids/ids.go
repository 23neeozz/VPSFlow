package ids

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// New generates a prefixed ULID identifier.
func New(prefix string) (string, error) {
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy)
	if prefix == "" {
		return id.String(), nil
	}
	return fmt.Sprintf("%s_%s", prefix, id.String()), nil
}
