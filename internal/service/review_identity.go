package service

import (
	"crypto/rand"
	"encoding/base32"
	"time"
)

type ReviewClock struct{}

func (ReviewClock) Now() time.Time { return time.Now().UTC() }

// ReviewIdentitySequence emits an opaque review-room identity.  The entropy
// avoids collisions between parallel coordinator processes and keeps IDs free
// from user or material data.
type ReviewIdentitySequence struct{}

func (*ReviewIdentitySequence) NewID() string {
	var entropy [10]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return "review-fallback-" + time.Now().UTC().Format("20060102T150405.000000000")
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(entropy[:])
	return "review-" + encoded
}
