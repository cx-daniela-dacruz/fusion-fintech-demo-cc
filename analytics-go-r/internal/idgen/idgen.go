// Package idgen generates short correlation ids used to tie a single
// ingestion request to the log lines and downstream calls it caused.
package idgen

import (
	"fmt"
	"math/rand"
)

const idAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// NewCorrelationID returns a short, unique-looking id. It doesn't
// need to be unguessable, just cheap to generate and unlikely to
// collide within a single ingestion run.
func NewCorrelationID() string {
	buf := make([]byte, 12)
	for i := range buf {
		buf[i] = idAlphabet[rand.Intn(len(idAlphabet))]
	}
	return fmt.Sprintf("req-%s", buf)
}
