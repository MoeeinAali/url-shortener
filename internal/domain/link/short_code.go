package link

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	shortCodeLength   = 7
	shortCodeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // base62
)

// ShortCode is a value object for the public short identifier of a link.
// It is immutable and always matches the format invariant (fixed length, base62).
type ShortCode struct {
	value string
}

// NewShortCode validates an existing short code string (e.g. read from storage
// or a request path).
func NewShortCode(raw string) (ShortCode, error) {
	if len(raw) != shortCodeLength {
		return ShortCode{}, ErrInvalidShortCode
	}
	for _, r := range raw {
		if !strings.ContainsRune(shortCodeAlphabet, r) {
			return ShortCode{}, ErrInvalidShortCode
		}
	}
	return ShortCode{value: raw}, nil
}

// GenerateShortCode produces a cryptographically-random base62 short code.
func GenerateShortCode() (ShortCode, error) {
	buf := make([]byte, shortCodeLength)
	max := big.NewInt(int64(len(shortCodeAlphabet)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return ShortCode{}, err
		}
		buf[i] = shortCodeAlphabet[n.Int64()]
	}
	return ShortCode{value: string(buf)}, nil
}

// String returns the short code text.
func (c ShortCode) String() string { return c.value }

// IsZero reports whether the value object is the zero value.
func (c ShortCode) IsZero() bool { return c.value == "" }

// Generator is the domain's default short-code generation strategy. It satisfies
// the application's ShortCodeGenerator port and keeps generation logic inside the
// domain while remaining injectable for tests.
type Generator struct{}

// Generate returns a new random short code.
func (Generator) Generate() (ShortCode, error) { return GenerateShortCode() }
