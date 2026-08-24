package provider

import (
	"crypto/rand"
	"math/big"
)

// testRandomSuffix returns a random non-negative integer for use in
// acceptance test resource names, so concurrent test runs don't collide on
// uniqueness constraints (e.g. Event Webhook URLs, subuser usernames).
func testRandomSuffix() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	if err != nil {
		// crypto/rand failing is effectively impossible on supported
		// platforms; a fixed fallback keeps tests from panicking.
		return 1
	}
	return n.Int64()
}
