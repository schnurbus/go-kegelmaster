package oldmigrate

import (
	"fmt"

	"github.com/google/uuid"
)

// Namespace for deterministic UUIDs (fixed value for reproducible migrations).
var namespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

// ID returns a deterministic UUID from entity type and old bigint ID.
func ID(entity string, oldID int64) uuid.UUID {
	return uuid.NewSHA1(namespace, []byte(fmt.Sprintf("%s:%d", entity, oldID)))
}

// IDFromComposite returns a deterministic UUID from entity and composite key parts.
func IDFromComposite(entity string, parts ...int64) uuid.UUID {
	b := []byte(entity)
	for _, p := range parts {
		b = append(b, fmt.Sprintf(":%d", p)...)
	}
	return uuid.NewSHA1(namespace, b)
}

// IDFromStrings returns a deterministic UUID from entity and string parts (e.g. for matchday_id:player_id).
func IDFromStrings(entity string, parts ...string) uuid.UUID {
	b := []byte(entity)
	for _, p := range parts {
		b = append(b, ':')
		b = append(b, p...)
	}
	return uuid.NewSHA1(namespace, b)
}
