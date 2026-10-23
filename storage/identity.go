package storage

import (
	"crypto/rand"
	"fmt"
)

// uuidVersion4 and uuidVariant are the bits that mark a UUID as random (RFC 9562).
const (
	uuidVersion4 = 0x40
	uuidVariant  = 0x80
)

// NewID makes a random UUID ("3f2b…-…"). Every row gets one on the device, so a later sync between
// devices never has to renumber anything.
func NewID() string {
	var bytes [16]byte
	rand.Read(bytes[:]) // crypto/rand never fails on the systems Go supports
	bytes[6] = bytes[6]&0x0f | uuidVersion4
	bytes[8] = bytes[8]&0x3f | uuidVariant
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
