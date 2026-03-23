package domain

import (
    "time"

    "github.com/google/uuid"
)

// RefreshToken is an opaque refresh token reference stored hashed at rest.
type RefreshToken struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    TokenHash  string
    ExpiresAt  time.Time
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Version    int64
    DeletedAt  *time.Time
}
